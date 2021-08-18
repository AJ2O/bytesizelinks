// construct DynamoDB connection
const AWS = require("aws-sdk");
const dynamodb = new AWS.DynamoDB();

// import URL validity module
const validUrl = require("valid-url");

// import tensorflow toxicity module
require("@tensorflow/tfjs");
const toxicity = require("@tensorflow-models/toxicity");
// The minimum prediction confidence.
const toxicityThreshold = 0.9;

// get environment information
const ALPHABET = process.env.alphabet;
const idGenerationRetries = parseInt(process.env.idGenerationRetries);
const idLength = parseInt(process.env.randomLinkLength);
const idKey = process.env.idKey;
const maxLinkLength = process.env.maxLinkLength;
const minLinkLength = process.env.minLinkLength;
const sourceLinkAttr = process.env.sourceLinkAttr;
const tableName = process.env.tableName;
const bslURL = process.env.bslURL;

// link generation helpers
function generateID() {
  var id = "";
  for (var i = 0; i < idLength; i++) {
    var randomIndex = Math.floor(Math.random() * ALPHABET.length);
    var randomChar = ALPHABET.charAt(randomIndex);
    id += randomChar;
  }
  return id;
}
async function generateUniqueID() {
  for (var i = 0; i < idGenerationRetries; i++) {
    var id = generateID();

    // check table if the id is already in-use
    var response = await doesLinkExist(id);
    // if it's not, it can be used
    if (!response) {
      return id;
    }
  }
  var error = new Error("Failed to generate a unique link!");
  throw error;
}

// custom link helpers
async function validateRequest(requestLink) {
  // maximum link length
  if (requestLink.length > maxLinkLength) {
    throw new Error(
      "The link should be at maximum " + maxLinkLength + " characters!"
    );
  }

  // minimum link length
  if (requestLink.length < minLinkLength) {
    throw new Error(
      "The link should be at minimum " + minLinkLength + " characters!"
    );
  }

  // check characters
  for (var letter of requestLink) {
    if (!ALPHABET.includes(letter)) {
      throw new Error(
        "Invalid request character: '" +
          letter +
          "'. Only numbers and letters can be used."
      );
    }
  }

  // check appropriateness of requested link
  var model = await toxicity.load(toxicityThreshold);
  var predictions = await model.classify([requestLink]);
  for (var prediction of predictions) {
    if (prediction["results"][0]["match"] == true) {
      throw new Error("Please choose a more appropriate custom link.");
    }
  }

  return true;
}
async function doesLinkExist(byteLink) {
  var itemKey = {};
  itemKey[idKey] = { S: byteLink };
  var params = {
    TableName: tableName,
    Key: itemKey,
  };

  var response = await dynamodb.getItem(params).promise();
  // if the response is empty, the id is available for use
  if (!Object.keys(response).length) {
    return false;
  }
  return true;
}

// table helpers
async function addByteLink(sourceURL, byteLink) {
  var newItem = {};
  newItem[idKey] = { S: byteLink };
  newItem[sourceLinkAttr] = { S: sourceURL };
  newItem["uses"] = { N: "0" };
  newItem["timesRequested"] = { N: "1" };
  var params = {
    TableName: tableName,
    Item: newItem,
  };
  try {
    await dynamodb.putItem(params).promise();
    var response = {
      statusCode: 200,
      body: bslURL + "/" + byteLink,
    };
    return response;
  } catch (err) {
    return {
      statusCode: 418,
      body: err.message,
    };
  }
}

// synchronous handler
exports.handler = (event, context, callback) => {
  // TODO: verify legitimacy of source link
  console.log("Request: " + JSON.stringify(event));
  var sourceLink = event["queryStringParameters"]["sourceLink"];

  // no empty source link
  if (!sourceLink || 0 === sourceLink.length) {
    callback(null, {
      statusCode: 400,
      body: "Parameter 'sourceLink' must be present and non-empty!",
    });
    return;
  }

  // the source link must look like a proper link
  if (!validUrl.isUri(sourceLink)) {
    callback(null, {
      statusCode: 400,
      body: sourceLink + " is not a valid URL!",
    });
    return;
  }

  // check for custom byte link
  var requestLink = event["queryStringParameters"]["customByteLink"];

  // if no custom byte link, generate a random one
  if (!requestLink || 0 === requestLink.length) {
    // generate random byte link
    generateUniqueID().then(
      // success
      function (value) {
        addByteLink(sourceLink, value).then(function (response) {
          callback(null, response);
        });
      },

      // error
      function (err) {
        console.log(err);
        callback(null, {
          statusCode: 418,
          body: err.message,
        });
      }
    );
  } else {
    // validate custom link
    // set to lowercase
    requestLink = requestLink.toLowerCase();

    // 1. Check validity of custom link
    validateRequest(requestLink).then(
      // success
      function (isValid) {
        // 2. Check if taken
        doesLinkExist(requestLink).then(
          // success
          function (value) {
            if (!value) {
              addByteLink(sourceLink, requestLink).then(function (response) {
                callback(null, response);
              });
            } else {
              callback(null, {
                statusCode: 409,
                body:
                  "The custom link: '" + requestLink + "' is currently taken!",
              });

              // update times requested
              var itemKey = {};
              itemKey[idKey] = { S: requestLink };
              var params = {
                TableName: tableName,
                Key: itemKey,
              };

              var updateParams = {
                ExpressionAttributeNames: {
                  "#T": "timesRequested",
                },
                ExpressionAttributeValues: {
                  ":t": {
                    N: "1",
                  },
                },
                TableName: tableName,
                Key: itemKey,
                UpdateExpression: "ADD #T :t",
              };
              dynamodb.updateItem(
                updateParams,
                function (updateErr, updateData) {
                  if (updateErr) {
                    console.log(updateErr);
                  }
                }
              );
            }
          },

          // error
          function (err) {
            console.log(err);
            callback(null, {
              statusCode: 418,
              body: err.message,
            });
          }
        );
      },

      // error
      function (err) {
        callback(null, {
          statusCode: 400,
          body: err.message,
        });
      }
    );
  }
};
