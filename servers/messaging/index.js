//@ts-check
"use strict";

const mongodb = require("mongodb");
const amqp = require("amqplib");
const MongoStore = require("./mongostore.js");
const express = require("express");
const morgan = require("morgan");
const moment = require("moment");

const addr = process.env.ADDR || ":80";
const [host, port] = addr.split(":");
const portNum = parseInt(port);

const mongoAddr = process.env.DBADDR || "mymongo:27017";
const mongoURL = `mongodb://${mongoAddr}/mgo`;

const qName = "testQ";
const mqAddr = process.env.MQADDR || "devrabbit:5672";
const mqURL = `amqp://${mqAddr}`;

const app = express();

const contentType = "Content-Type";
const typeJSON = "application/json";
const typeText = "text/plain";

function headerCheck(req, res) {
    let userJSON = req.get("X-User");
    if (!userJSON) {
        throw Error("No X-User header provided");
    }
}

(async function() { 
    try {
        console.log("connecting to %s", mqURL);
        let connection = await amqp.connect(mqURL);
        let ch = await connection.createChannel();
        let qConf = await ch.assertQueue(qName, {durable: false});

        console.log("starting to send messages...");

        mongodb.MongoClient.connect(mongoURL)
        .then(db => {
            let messageStore = new MongoStore(db, "messages");
            let channelStore = new MongoStore(db, "channels");

            app.use(express.json());
            app.use(morgan(process.env.LOG_FORMAT || "dev"));

            channelStore.getByName("general")
                .then(general => {
                    if (!general) {
                        channelStore.insert({name: 'general'})
                    }
                })
                .catch(err => {
                    throw err
                });

            app.use("/v1/channels/:chanID", (req, res, next) => {
                headerCheck(req, res);
                res.set(contentType, typeJSON)

                let chanID = new mongodb.ObjectID(req.params.chanID);

                channelStore.get(chanID)
                    .then(channel => {
                        if (!channel) {
                            res.status(400);
                            next(Error("channel does not exist"));
                        }
                        if (req.method == "PATCH" || req.method == "DELETE") {
                            if (channel.creator != req.get("X-User")) {
                                res.status(403);
                                next(Error(`expected ${channel.creator}\n but got ${req.get("X-User")}`));
                            }
                        }
                    })
                    .catch(err => {
                        throw err;
                    });

                switch (req.method) {
                    case "GET":
                        //Respond with the latest 50 messages posted to the 
                        //specified channel, encoded as a JSON array
                        messageStore.getByID(req.params.chanID, 50)
                            .then(messages => {
                                res.json(messages);
                            })
                            .catch(err => {
                                throw err;
                            });
                        break;
                    case "POST":
                        //Create a new message in this channel using the JSON in the request body,
                        //and respond with a copy of the new message. The only message property you should read
                        //from the request is body. Set the others based on context.
                        let message = {
                            channelID: req.params.chanID,
                            body: req.body.body,
                            createdAt: moment().format('MMMM Do YYYY, h:mm:ss a'),
                            creator: req.get("X-User"),
                            editedAt: ""
                        }

                        messageStore.insert(message)
                            .then(message => {
                                let event = {
                                    type: "message-new",
                                    message: message
                                }
                                ch.sendToQueue(qName, Buffer.from(JSON.stringify(event)));
                                res.json(message);
                            })
                            .catch(err => {
                                throw err;
                            });
                        break;
                    case "PATCH":
                        //If the current user created the channel, update only the name and/or description using the JSON
                        //in the request body and respond with a copy of the newly-updated channel encoded as a JSON object.
                        //If the current user isn't the creator, respond with the status code 403 (Forbidden).
                        channelStore.get(chanID)
                            .then(channel => {
                                let updates = {
                                    name: req.body.name || channel.name,
                                    description: req.body.description || channel.description,
                                    editedAt: moment().format('MMMM Do YYYY, h:mm:ss a')
                                }
                                channelStore.update(chanID, updates)
                                    .then(updated => {
                                        let event = {
                                            type: "channel-update",
                                            channel: updated
                                        }
                                        ch.sendToQueue(qName, Buffer.from(JSON.stringify(event)));
                                        res.json(updated);
                                    })
                                    .catch(err => {
                                        throw err;
                                    });
                            })
                            .catch(err => {
                                throw err;
                            });
                        break;
                    case "DELETE":
                        //If the current user created the channel, delete it and all messages related to it.
                        //If the current user isn't the creator, respond with the status code 403 (Forbidden).
                        messageStore.deleteAll(req.params.chanID)
                            .then(deleted => {
                                channelStore.delete(chanID)
                                    .catch(err => {
                                        throw err;
                                    });
                                let event = {
                                    type: "channel-delete",
                                    channelID: chanID
                                }
                                ch.sendToQueue(qName, Buffer.from(JSON.stringify(event)));
                                res.json(deleted);
                            })
                            .catch(err => {
                                throw err;
                            });
                        break;
                    default:
                        res.status(405);
                        throw new Error("method must be GET, POST, PATCH or DELETE");
                    
                }
            });

            app.use("/v1/channels", (req, res, next) => {
                headerCheck(req, res);
                res.set(contentType, typeJSON);

                switch (req.method) {
                    case "POST":
                        //Create a new channel using the JSON in the request body and 
                        //respond with a copy of the new channel encoded as a JSON object. 
                        //The name property is required, but description is optional.
                        let json = req.body;
                        let channelName = json.name;
                        if (!channelName) {
                            res.status(400);
                            next(Error("channel name is mandatory"));
                        }

                        channelStore.getByName(channelName)
                            .then(channel => {
                                if (channel) {
                                    res.status(400);
                                    next(Error("channel must have a unique name"));
                                }
                            })
                            .catch(err => {
                                throw err;
                            });

                        let channel = {
                            name: channelName,
                            description: json.description || "",
                            createdAt: moment().format('MMMM Do YYYY, h:mm:ss a'),
                            creator: req.get("X-User"),
                            editedAt: ""
                        }

                        channelStore.insert(channel)
                            .then(channel => {
                            let event = {
                                type: "channel-new",
                                channel: channel
                            }
                            ch.sendToQueue(qName, Buffer.from(JSON.stringify(event)));
                            res.json(channel);
                            })
                            .catch(err => {
                                throw err;
                            });
                        break;
                    case "GET":
                        //Respond with the list of all channels (just the channel data itself,
                        //not the messages), encoded as a JSON array.
                        channelStore.getAll()
                            .then(channels => {
                                res.json(channels);
                            })
                            .catch(err => {
                                throw(err)
                            });
                        break;
                    default:
                        res.status(405);
                        throw new Error("method must be GET or POST");
                }
            });

            app.use("/v1/messages/:messageID", (req, res, next) => {
                headerCheck(req, res);

                let messageID = new mongodb.ObjectID(req.params.messageID);
                messageStore.get(messageID)
                    .then(message => {
                        if (!message) {
                            res.status(400);
                            next(Error("no message found"));
                        }
                        if (message.creator != req.get("X-User")) {
                            res.status(403);
                            next(Error("you did not create this message"));
                        }
                    })
                    .catch(err => {
                        throw err;
                    });

                switch (req.method) {
                    case "PATCH":
                        //If the current user created the message, update the message body property using the JSON
                        //in the request body, and respond with a copy of the newly-updated message.
                        //If the user is not authenticated or is not the creator, respond with the status code 403 (Forbidden).
                        res.set(contentType, typeJSON);
                            let updates = {
                                body: req.body.body,
                                editedAt: moment().format('MMMM Do YYYY, h:mm:ss a')
                            }
                            messageStore.update(messageID, updates)
                                .then(message => {
                                    let event = {
                                        type: "message-update",
                                        message: message
                                    }
                                    ch.sendToQueue(qName, Buffer.from(JSON.stringify(event)));
                                    res.json(message);
                                })
                                .catch(err => {
                                    throw err
                                });
                        break;

                    case "DELETE":
                        //If the current user created the message, delete it and respond with a the text "message deleted".
                        //If the current user is not the creator, respond with the status code 403 (Forbidden).
                        res.set(contentType, typeText);
                            messageStore.delete(messageID)
                                .then(message => {
                                    let event = {
                                        type: "message delete",
                                        messageID: messageID
                                    }
                                    ch.sendToQueue(qName, Buffer.from(JSON.stringify(event)));
                                    res.send("message deleted");
                                })
                                .catch(err => {
                                    throw err
                                });
                        break;

                    default:
                        res.status(405);
                        throw new Error("method must be PATCH or DELETE");
                }
            });

            app.use((err, req, res, next) => {
                console.error(err.stack)
                res.set(contentType, typeText);

                if (err.message == "No X-User header provided") {
                    res.status(401).send(err.message);
                } else {
                    if (!res.status) {
                        res.status(500);
                    }
                    res.send(err.message);
                }
            });

            app.listen(portNum, host, () => {
                console.log(`server is listening at http://${addr}...`);
            });
        });
    } catch (err) {
        console.error(err.stack);
    }
})();
