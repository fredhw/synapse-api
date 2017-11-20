"use strict";

class Channel {
    constructor(props) {
        Object.assign(this, props)
    }

    validate() {
        if (!this.name) {
            return new Error('you must supply a name')
        }
    }
}

module.exports = Channel;