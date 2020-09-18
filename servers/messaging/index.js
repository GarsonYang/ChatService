const mongoose = require('mongoose');
const express = require('express');
const amqp = require('amqplib/callback_api');
const {channelSchema, messageSchema} = require('./schemas');
const {channelsHandler, channelMembersHandler, specificChannelHandler, specificMessageHandler} = require('./handlers')

// RabbitMQ
let rabbitChannel;

const getRabbitChannel = () => {
    return rabbitChannel;
}


const mongoEndpoint = "mongodb://mongodb:27017/messaging";
var port = process.env.PORT;
if (!port) {
    port = '4001';
}

const Channel = mongoose.model("Channel", channelSchema);
const Message = mongoose.model("Message", messageSchema);
// Channel.remove({}, function(err) { 
//     console.log('collection removed') 
// });

const app = express();
app.use(express.json());
app.enable('strict routing');

const connect = () => {
    mongoose.connect(mongoEndpoint);
}

const RequestWrapper = (handler, SchemaAndDbForwarder) => {
    return (req, res) => {
        handler(req, res, SchemaAndDbForwarder);
    }
}

app.use('/v1/channels/:channelID/members', RequestWrapper(channelMembersHandler, {Channel}));
app.use('/v1/channels/:channelID', RequestWrapper(specificChannelHandler, {Channel, Message, getRabbitChannel}));
app.use('/v1/channels', RequestWrapper(channelsHandler, {Channel, getRabbitChannel}));
app.use('/v1/messages/:messageID', RequestWrapper(specificMessageHandler, {Channel, Message, getRabbitChannel}));

connect();
mongoose.connection.on('error', console.error)
//    .on('disconnected', connect)
    .once('open', main);

async function main() {
    Channel.count({}, function(err, count) {
        if(count == 0) {
            var createdAt = new Date();
            var channel = {
                name: "general",
                private: false,
                creator: {},
                createdAt: createdAt,
                editedAt: createdAt
            };

            var query = new Channel(channel);
            query.save();
        }
    });
    
    amqp.connect('amqp://rabbitmq:5672', function(error0, connection) {
        if (error0) {
            console.log("Error connecting to rabbit instance");
            process.exit(1);
        }

        connection.createChannel(function(error1, ch) {
            if (error1) {
                console.log("Error creating channel");
                process.exit(1);
            }
    
            ch.assertQueue("events", {
                durable: true
            });
            rabbitChannel = ch;

            // ch.consume("events", (msg)=> {
            //     console.log("From rabbit:" + msg.content.toString())
            // }, {
            //     noAck: true
            // });
             

            app.listen(port, "", () => {
                console.log(`server listening ${port}`);
            });
        });
    });


}