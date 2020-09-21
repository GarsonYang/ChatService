# **Chat Service**

 This is a web chat application containing both front end and back end. On the back end, the architecture is carefully designed. It implements the API-gateway architecture and microservices are built.
  ## Application Architecture
  ![Image of API gateway Arch](https://i.ibb.co/4NGc4pn/api-gateway.jpg)
  
  I implemented the API-gateway architecture in the design. All the http requests coming from the clients go to the gateway and will be further dispatched to different microservices. The general sign-up and sign-in info would be sent to related microservices and get processed and responded, including the authentication sessions. 
  
  Messaging related requests would first go to the gateway, trigger a new web socket construction, go to the messaging microservice, get processed and then trigger a new Message Queue event which will be sent back to the API gateway. The API gateway will do something according to the instruction in the event. (More details explained below) 
  
  Compared to the monolithic architecture, the system become super scalable and easy to maintain. By this design, if anything goes wrong with one feature, the rest of the system will not be affected at all. Also, if a new feature is to be added to the back end, I can just work on the new microservice solely without touching any other existing parts. Once it's done, I can simply connect it to the multiplexer, pretty much like a "plug-in".
  
  ## Tech Stack
  - ### **Programming Language**
    The API gateway is written in **Golang**, which is a super convenient and capable programing language that I would recommend to everyone who work on the back end. For the microservices, most of them are written in Golang as well, but I used **Node.js** for the messaging microservice.
  - ### **Database**
    A **MySql** database is used to store the user information, including hashed passwords, considering those are mostly highly structured data.
    For sessions and authentications, I deployed a **Redis** database, for faster storage and access, as well as its ephemeral storage because the sessions would expire by design.
    For the messages and channels, a **MongoDB** is used. The data could be very loosely structured. For every message and channel, I need to store the users who have the access to see it and the users who have the access to administer it. Therefore, a NoSql database would be a better option here.
  ## Key Design
  - ### **Web Socket**
    To show live chat message update without forcing the user to refresh the page, web socket is used here to connect the API gateway and the clients.
  - ### **Multithread**
    Because we expect multiple clients to connect and use the chat room at the same time, I need to get ready for creating multiple web sockets and let all of them stay alive. Go provides a very convenient way to enable multithreading, which is called "go func". It will set a new thread every time a new web socket connection is established, and the web socket will be running on that thread.
  - ### **Message Queue**
    In my design, the messaging microservice receive the http request from the API gateway, but it won't directly response to the request, because it needs to response through the web socket connection, which is located in the gateway. Therefore, I need to send the response back to the gateway from the messaging microservice. It's a simple, one-way communication, so I set a RabbitMQ.
    Message Queue is like a producer-consumer pipeline. An event is generated on one end and gets consumed on the other end. Here, the messaging microservice will produce an event (the response) push it into the MQ. On the other end, the gateway will consume the event and do the job according to the instruction in the event.
  - ### **User-Connection Map**
    Since the API gateway get events from the MQ and then need to do some work through the web socket connection, how can the gateway know which web socket connection the event is related to? The answer is that I created a map to store the live web socket connections and their corresponding users when the connection was first built. Therefore, when the messaging microservice sends back the event through the MQ, I can process it over the correct web socket connection and then further send it to the right client.
  - ### **Load Balance**
    The amount of users and messages could be in totally different level. There could huge amount of messaging requests without too many users. Thus, I need to figure out a way to work around heavy loads. By our design, increasing the messaging capacity is not too hard. I can simply create a new messaging service on a new server with the exact same content as the original one and set the HTTP routing, and then I need to set some load balancing mechanism. Here I implement the round-robin technique, which means I try to evenly dispatch the requests to all available servers.
  - ### **Locks**
    For the User-Connection map and the load balancing counter, there could be times when two or more threads are accessing the data structure or variable at the same time. Therefore, I need to set locks to limit the access. Go provides good libraries called mutex that will do the exact work so that threads can only access the data structure or variable one by one.
  - ### **Trie**
    The trie data structure is implemented to realize super fast user search.
  - ### **Password Classification**
    All password is hashed using SHA256 before storage.
