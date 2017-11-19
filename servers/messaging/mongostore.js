"use strict";

const mongodb = require("mongodb");

class MongoStore {
    constructor(db, colName) {
        this.collection = db.collection(colName);
    }
    insert(object) {
        object._id = new mongodb.ObjectID();
        return this.collection.insertOne(object)
            .then(() => object);
    }
    update(id, updates) {
        let updateDoc = {
            "$set": updates
        }
        return this.collection.findOneAndUpdate(
            {_id: id}, 
            updateDoc, 
            {returnOriginal: false})
            .then(result => result.value);
    }
    get(id) {
        return this.collection.findOne({_id: id});
    }
    getByName(channelName) {
        return this.collection.findOne({name: channelName});
    }
    delete(id) {
        return this.collection.deleteOne({_id: id});
    }
    deleteAll(id) {
        return this.collection.deleteMany({channelID: id});
    }
    getByID(id, limit) {
        return this.collection.find({channelID: id})
            .limit(limit)
            .toArray();
    }
    getAll() {
        return this.collection.find({}).toArray();
    }
}

module.exports = MongoStore;