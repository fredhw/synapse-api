//@ts-check
"use strict";

const express = require("express");

module.exports = (mongoSession) => {
    if (!mongoSession) {
        throw new Error("provide Mongo session");
    }
    let router = express.Router();
    router.get("/v1/channels", (req, res) => {
        res.json([{name: "general"}]);
    });

    return router;
}