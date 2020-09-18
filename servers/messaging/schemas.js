const Schema = require('mongoose').Schema;

const channelSchema = new Schema({
    id: {type: Schema.Types.ObjectId, auto: true},
    name: {type: String, required: true, unique: true},
    description: {type: String, required: false},
    private: {type: Boolean, required: true},
    members: [],
    createdAt: {type:Date, required: true},
    creator: {type: JSON, required:true},
    editedAt: {type:Date}
});

const messageSchema = new Schema({
    id: {type: Schema.Types.ObjectId, auto: true},
    channelID: {type: Schema.Types.ObjectId, required: true},
    body: {type: String},
    createdAt: {type:Date, required: true},
    creator: {type: JSON, required:true},
    editedAt: {type:Date}
});

module.exports = {channelSchema, messageSchema};