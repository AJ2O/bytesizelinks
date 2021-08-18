// construct DynamoDB connection
const AWS = require("aws-sdk");
const dynamodb = new AWS.DynamoDB();

// get environment information
const ALPHABET = process.env.alphabet;
const idKey = process.env.idKey;
const sourceLinkAttr = process.env.sourceLinkAttr;
const tableName = process.env.tableName;

// link validity helpers
function validateLink(byteLink) {
  for (var letter of byteLink) {
    if (!ALPHABET.includes(letter)) {
      var error = new Error(
        "Invalid request character: '" +
          letter +
          "'. Only numbers and letters can be used."
      );
      throw error;
    }
  }
  return true;
}

exports.handler = (event, context, callback) => {
  // TODO: verify legitimacy of source link
  var byteLink = event["queryStringParameters"]["byteLink"];

  // no empty parameters
  if (!byteLink || 0 === byteLink.length) {
    callback(null, {
      statusCode: 400,
      body: "Parameter 'byteLink' must be present and non-empty!",
    });
    return;
  }

  // validate format of the link
  try {
    validateLink(byteLink);
  } catch (err) {
    callback(null, {
      statusCode: 400,
      body: err.message,
    });
    return;
  }

  // return original link
  var itemKey = {};
  itemKey[idKey] = { S: byteLink };
  var params = {
    TableName: tableName,
    Key: itemKey,
  };
  dynamodb.getItem(params, function (err, data) {
    if (err) {
      callback(null, {
        statusCode: 418,
        body: err.message,
      });
    } else {
      // if the data is empty, the link is invalid
      if (!Object.keys(data).length) {
        var response = {
          statusCode: 404,
          body: "This byte-link does not exist!",
        };
        callback(null, response);
      } else {
        var response = {
          statusCode: 200,
          body: data["Item"][sourceLinkAttr]["S"],
        };
        callback(null, response);

        // update item uses
        var updateParams = {
          ExpressionAttributeNames: {
            "#U": "uses",
          },
          ExpressionAttributeValues: {
            ":u": {
              N: "1",
            },
          },
          TableName: tableName,
          Key: itemKey,
          UpdateExpression: "ADD #U :u",
        };
        dynamodb.updateItem(updateParams, function (updateErr, updateData) {
          if (updateErr) {
            console.log(updateErr);
          }
        });
      }
    }
  });
};
