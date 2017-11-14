"use strict";

const MongoStore = require("./messagestore");
const mongodb = require("mongodb");

const mongoAddr = process.env.DBADDR || "localhost:27017";
const mongoURL = `mongodb://${mongoAddr}/messages`;

describe("Mongo Message Store", () => {
    test("CRUD Cycle", () => {
        return mongodb.MongoClient.connect(mongoURL)
            .then(db => {
                let store = new MongoStore(db, "messages");
                let task = {
                    title: "Learn Node.js to MongoDB",
                    tags: ["mongodb", "node.js", "info344"]
                };
                let taskID;

                return store.insert(task)
                    .then(task => {
                        expect(task._id).toBeDefined();
                        taskID = task._id;
                        return task._id;
                    })
                    .then(taskId => {
                        return store.get(taskId);
                    })
                    .then(fetchedTask => {
                        expect(fetchedTask).toEqual(task);
                        return store.update(task._id, {completed: true});
                    })
                    .then(updatedTask => {
                        expect(updatedTask.completed).toBe(true);
                        return store.delete(task._id);
                    })
                    .then(() => {
                        return store.get(task._id);
                    })
                    .then(fetchedTask => {
                        expect(fetchedTask).toBeFalsy();
                    })
                    .then(() => {
                        db.close();
                    })
                    .catch(err => {
                        db.close();
                        throw err;
                    });
            });
    });
});