//@ts-check
"use strict";

const mongodb = require("mongodb");
const MongoStore = require("./messageStore");
const express = require("express");

const addr = process.env.ADDR || ":80";
const [host, port] = addr.split(":");
const portNum = parseInt(port);

const mongoAddr = process.env.DBADDR || "localhost:27017";
const mongoURL = `mongodb://${mongoAddr}/tasks`;
const app = express();

mongodb.MongoClient.connect(mongoURL)
    .then(db => {
        let messageStore = new MongoStore(db, "messages");
        let channelStore = new MongoStore(db, "channels");

        app.use(express.json());


        //Create a new channel using the JSON in the request body and respond with a copy of the
        //new channel encoded as a JSON object. The name property is required, but description is optional.
        app.post("/v1/channels", (req, res) => {
            let userJSON = req.get("X-User");
            if (!userJSON) {
                res.status(401);
                throw new Error("No X-User header provided");
            }
        });

        //Respond with the list of all channels (just the channel data itself, not the messages),
        //encoded as a JSON array.
        app.get("/v1/channels", (req, res) => {
            let userJSON = req.get("X-User");
            if (!userJSON) {
                res.status(401);
                throw new Error("No X-User header provided");
            }
        });

        //Respond with the latest 50 messages posted to the specified channel, encoded as a JSON array
        app.get("/v1/channels/:chanID", (req, res) => {
            let userJSON = req.get("X-User");
            if (!userJSON) {
                res.status(401);
                throw new Error("No X-User header provided");
            }
        });

        //Create a new message in this channel using the JSON in the request body,
        //and respond with a copy of the new message. The only message property you should read
        //from the request is body. Set the others based on context.
        app.post("/v1/channels/:chanID", (req, res) => {
            let userJSON = req.get("X-User");
            if (!userJSON) {
                res.status(401);
                throw new Error("No X-User header provided");
            }
        });

        //If the current user created the channel, update only the name and/or description using the JSON
        //in the request body and respond with a copy of the newly-updated channel encoded as a JSON object.
        //If the current user isn't the creator, respond with the status code 403 (Forbidden).
        app.patch("/v1/channels/:chanID", (req, res) => {
            let userJSON = req.get("X-User");
            if (!userJSON) {
                res.status(401);
                throw new Error("No X-User header provided");
            }
        });

        //If the current user created the channel, delete it and all messages related to it.
        //If the current user isn't the creator, respond with the status code 403 (Forbidden).
        app.delete("/v1/channels/:chanID", (req, res) => {
            let userJSON = req.get("X-User");
            if (!userJSON) {
                res.status(401);
                throw new Error("No X-User header provided");
            }
        });

        //If the current user created the message, update the message body property using the JSON
        //in the request body, and respond with a copy of the newly-updated message.
        //If the user is not authenticated or is not the creator, respond with the status code 403 (Forbidden).
        app.get("/v1/messages/:messageID", (req, res) => {
            let userJSON = req.get("X-User");
            if (!userJSON) {
                res.status(401);
                throw new Error("No X-User header provided");
            }
        });

        //If the current user created the message, delete it and respond with a the text "message deleted".
        //If the current user is not the creator, respond with the status code 403 (Forbidden).
        app.delete("/v1/messages/:messageID", (req, res) => {
            let userJSON = req.get("X-User");
            if (!userJSON) {
                res.status(401);
                throw new Error("No X-User header provided");
            }
        });

        app.listen(portNum, host, () => {
            console.log(`server is listening at http://${addr}...`);
        });
    })
    .catch(err => {
        console.error(err.stack);
        throw err;
    });
