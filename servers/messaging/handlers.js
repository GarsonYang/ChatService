const channelsHandler = async (req, res, {Channel, getRabbitChannel}) => {
    if (!req.headers['x-user']) {
        res.status(401).send("unauthorized, log in first");
        return;
    }

    switch (req.method){
        case "GET":
            Channel.find((err, channels) => {
                if (err) {
                    res.status(500).send('Unable to get channels');
                    return;
                }

                var filtered = [];
                channels.forEach(channel => {
                    if(channel.private == true ){
                        var members = channel.members;
                        
                        members.some(member => {
                            if(member.id == JSON.parse(req.headers["x-user"]).id) {
                                filtered.push(channel);
                                return true;
                            }
                        });
                    } else {
                        filtered.push(channel);
                    }
                });

                res.setHeader("Content-Type", "application/json");
                res.json(filtered);
            });

            break;

        case "POST":
            var {name, description, private, members} = req.body;
            if(!name) {
                res.status(400).send("Channel name cannot be empty");
                return;
            }

            if (private == true && !members) {
                res.status(400).send("Members must be specified for private channel");
                return;
            }

            var createdAt = new Date();
            var editedAt = new Date();
            console.log(req.headers["x-user"]);
            var creator = JSON.parse(req.headers["x-user"]);
            var channel = {
                name,
                private,
                createdAt,
                creator,
                editedAt,
                description,
                members
            };
            
            var query = new Channel(channel);
            query.save((err, newChannel) => {
                if (err) {
                    if(err.errmsg.startsWith("E1100")){
                        res.status(400).send("Channel name is taken. Please try another name");
                        return;
                    }
                    res.status(500).send(err.errmsg);
                    return;
                }
                
                let ch = getRabbitChannel();
                let memberIDs = [];
                newChannel.members.forEach(member => {
                    memberIDs.push(member.id);
                });
                ch.sendToQueue("events", Buffer.from(JSON.stringify(
                    {
                        type: "channel-new",
                        channel: newChannel,
                        userIDs: memberIDs
                    }
                )))

                res.setHeader("Content-Type", "application/json");
                res.status(201).json(newChannel);
            });
            

            break;

        default:
            res.status(405).send("Method not allowed");
    }
}

const channelMembersHandler = async (req, res, {Channel}) => {
    if (!req.headers['x-user']) {
        res.status(401).send("unauthorized, log in first");
        return;
    }
    const user = JSON.parse(req.headers['x-user']);
    
    const channelID = req.params["channelID"];
    Channel.findOne({id: channelID}, (err, channel) => {
        if (err) {
            res.status(500).send('Unable to get channels');
            return;
        }

        if(user.id != channel.creator.id) {
            res.status(403).send("Operation is forbidden: only the channel creator can manage the channel");
            return;
        }

        switch (req.method){
            case "POST":
                const newMember = req.body;
                try {
                    channel.members.push(newMember);
                    var editedAt = new Date();
                    channel.editedAt = editedAt;
                    channel.save();
                } catch (e) {
                    res.status(500).send(e);
                    return;
                }
                res.status(201).send("New channel member is added");
                return;
            case "DELETE":
                const memberToRemove = req.body;
                for(i=0; i<channel.members.length; i++){
                    if (channel.members[i].id == memberToRemove.id) {
                        channel.members = channel.members.slice(0,i).concat(channel.members.slice(i+1));
                        var editedAt = new Date();
                        channel.editedAt = editedAt;
                        channel.save();
                        res.status(200).send("The member is removed from the channel");
                        return;
                    }
                }
                res.status(400).send("No specified member found in the channel");
                return;
            default:
                res.status(405).send("Method not allowed");
        }

    });

}

const specificChannelHandler = async (req, res, {Channel, Message, getRabbitChannel}) => {
    if (!req.headers['x-user']) {
        res.status(401).send("unauthorized, log in first");
        return;
    }
    const user = JSON.parse(req.headers['x-user']);

    const channelID = req.params["channelID"];
    if(!channelID){
        res.status(401).send("no channel found");
        return;
    }
    Channel.findOne({id: channelID}, (err, channel) => {
        if (err) {
            res.status(500).send('Unable to get channels');
            return;
        }

        const checkIsMember = () => {
            if (channel.private == true) {
                var members = channel.members;
                var isMember = false;
                for(i=0; i<members.length; i++){
                    if(members[i].id == JSON.parse(req.headers["x-user"]).id) {
                        isMember = true;
                        break;
                    }
                }
                if (!isMember) {
                    res.status(403).send('Not a member of current Channel');
                    return false;
                }
            }
            return true;
        }

        let ch = getRabbitChannel();
        let memberIDs = [];
        channel.members.forEach(member => {
            memberIDs.push(member.id);
        });

        switch (req.method){
            case "GET":
                if(!checkIsMember()) return;
                Message.find({channelID: channelID}, (err, messages) => {
                    if (err) {
                        res.status(500).send('Unable to get messages');
                        return;
                    }

                    messages.sort((a,b) => {
                        return b.createdAt - a.createdAt;
                    })

                    const max = 100;
                    
                    const before = req.query.before;
                    if(before){
                        for(i=0; i<messages.length; i++) {
                            if(messages[i].id == before) {
                                res.setHeader("Content-Type", "application/json");
                                if(i==messages.length-1) {
                                    res.status(400).send('no more messages');
                                    return;
                                }
                                res.json(messages.slice(i+1, i+1+max));
                                return;
                            }
                        }
                        res.status(400).send('invalid messageID provided in "before"');
                        return;
                    }
                    res.setHeader("Content-Type", "application/json");
                    res.json(messages.slice(0, max));
                })
                return;
            case "POST":
                if(!checkIsMember()) return;
                var {body} = req.body;
                var createdAt = new Date();
                var editedAt = new Date();
                var creator = JSON.parse(req.headers["x-user"]);
                var message = {
                    channelID,
                    body,
                    createdAt,
                    creator,
                    editedAt,
                };

                var query = new Message(message);
                query.save((err, newMessage) => {
                    if (err) {
                        res.status(500).send(err);
                        return;
                    }
                    
                    // rabbitmq
                    ch.sendToQueue("events", Buffer.from(JSON.stringify(
                        {
                            type: "message-new",
                            message: newMessage,
                            userIDs: memberIDs
                        }
                    )))

                    res.setHeader("Content-Type", "application/json");
                    res.status(201).json(newMessage);
                });
                return;
            case "PATCH":
                if(user.id != channel.creator.id) {
                    res.status(403).send("Operation is forbidden: only the channel creator can manage the channel");
                    return;
                }

                const {name, description} = req.body;
                try {
                    channel.name = name;
                    channel.description = description;
                    var editedAt = new Date();
                    channel.editedAt = editedAt;
                    channel.save();
                } catch (e) {
                    res.status(500).send(e);
                    return;
                }

                // rabbitmq
                ch.sendToQueue("events", Buffer.from(JSON.stringify(
                    {
                        type: "channel-update",
                        channel: channel,
                        userIDs: memberIDs
                    }
                )))

                res.setHeader("Content-Type", "application/json");
                res.status(200).json(channel);
                return;
            case "DELETE":
                if(user.id != channel.creator.id) {
                    res.status(403).send("Operation is forbidden: only the channel creator can manage the channel");
                    return;
                }

                Channel.deleteOne({id:channelID}, (err) => {
                    if(err) res.status(500).send(err);

                    // rabbitmq
                    ch.sendToQueue("events", Buffer.from(JSON.stringify(
                        {
                            type: "channel-delete",
                            channelID: channelID,
                            userIDs: memberIDs
                        }
                    )))
                });

                res.status(200).send("The delete is successful");
                return;
            default:
                res.status(405).send("Method not allowed");
        }
    })
}

const specificMessageHandler = async (req, res, {Channel, Message, getRabbitChannel}) => {
    if (!req.headers['x-user']) {
        res.status(401).send("unauthorized, log in first");
        return;
    }
    const user = JSON.parse(req.headers['x-user']);

    const messageID = req.params["messageID"]
    Message.findOne({id: messageID}, async (err, message) => {
        if (err) {
            res.status(500).send('Unable to get message');
            return;
        }

        if(user.id != message.creator.id) {
            res.status(403).send("Operation is forbidden: only the message creator can edit the message");
            return;
        }

        let ch = getRabbitChannel();
        switch (req.method) {
            case "PATCH":
                var {body} = req.body;
                try {
                    message.body = body;
                    var editedAt = new Date();
                    message.editedAt = editedAt;
                    message.save();
                } catch (e) {
                    res.status(500).send(e);
                    return;
                }

                // rabbitmq
                Channel.findOne({id: message.channelID}, (err, channel) => {
                    if (err) {
                        res.status(500).send('Unable to find the channel of the message');
                        return;
                    }
                    let memberIDs = [];
                    channel.members.forEach( member => {
                        memberIDs.push(member.id);
                    });

                    ch.sendToQueue("events", Buffer.from(JSON.stringify(
                        {
                            type: "message-update",
                            message: message,
                            userIDs: memberIDs
                        }
                    )))
                });

                res.setHeader("Content-Type", "application/json");
                res.json(message);
                return;
                
            case "DELETE":
                let messageID = message.id;
                Message.deleteOne({id: message.id}, (err) => {
                    if (err) {
                        res.status(500).send(err);
                        return;
                    }
                    res.status(200).send("The delete is successful");
                })

                // rabbitmq
                Channel.findOne({id: message.channelID}, (err, channel) => {
                    if (err) {
                        res.status(500).send('Unable to find the channel of the message');
                        return;
                    }
                    let memberIDs = [];
                    channel.members.forEach( member => {
                        memberIDs.push(member.id);
                    });

                    ch.sendToQueue("events", Buffer.from(JSON.stringify(
                        {
                            type: "message-delete",
                            messageID: messageID,
                            userIDs: memberIDs
                        }
                    )))
                });

                return;
            default:
                res.status(405).send("Method not allowed");
        }
        
    })
    
}

module.exports = {channelsHandler, channelMembersHandler, specificChannelHandler, specificMessageHandler};