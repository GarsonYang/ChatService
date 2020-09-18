let state = {
    currentChannelID: ""
}

function handleError(errText) {
    $("#errMsg").text(errText);
}

function goToMainPage(auth, response) {
    setLocalStorage(auth, response)
    
    window.location.replace("main.html");
}

function setLocalStorage(auth, response){
    localStorage.setItem("Authorization", auth);

    var data = JSON.parse(response);
    localStorage.setItem("id", data.id);
    localStorage.setItem("userName", data.userName);
    localStorage.setItem("firstName", data.firstName);
    localStorage.setItem("lastName", data.lastName);
    localStorage.setItem("photoURL", data.photoURL);
}

function getChannels(){
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        if (this.readyState == 4) {
            if(this.status < 300){
                renderChannelList(this.responseText);
            }
            else {
                handleError(this.responseText);
            }
        }
    };

    xhttp.open("GET", "https://api.garson.me:443/v1/channels", true);
    xhttp.setRequestHeader("Content-Type", "application/json");
    xhttp.setRequestHeader("Authorization", localStorage.getItem("Authorization"));
    xhttp.send();
}

function creatChannel(name){
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        if (this.readyState == 4) {
            if(this.status < 300){
                getChannels();
            }
            else {
                alert(this.responseText);
            }
        }
    };
    var newChannel = {}
    newChannel["name"] = name;
    newChannel["private"] = false;

    xhttp.open("POST", "https://api.garson.me:443/v1/channels", true);
    xhttp.setRequestHeader("Content-Type", "application/json");
    xhttp.setRequestHeader("Authorization", localStorage.getItem("Authorization"));
    xhttp.send(JSON.stringify(newChannel));
}

function newChannelPrompt(){
    var name = prompt("Please enter the channel name:", "Channel 1");
    if (name == "") {
        alert("Input a valid name");
    } else if (name != null) {
        creatChannel(name);
    }
}

function renderChannelList(channelList){
    $("#channel-list").empty();
    var channels = JSON.parse(channelList);
    jQuery.each(channels, function() {
        saveChannelIdToLocal(this);

        let elem = $("#channel-list").append('<div class="sidebar-channel" id="channel-' + this.id + '">#' + this.name + '</div>');
        //elem.click(switchChannel);
    });
}

function saveChannelIdToLocal(channel){
    localStorage.setItem(channel.name, channel.id);
}

function deleteChannel(){
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        if (this.readyState == 4) {
            if(this.status < 300){
                renderChannelList(this.responseText);
            }
            else {
                handleError(this.responseText);
            }
        }
    };

    xhttp.open("DELETE", "https://api.garson.me:443/v1/channels" + currentChannelID, true);
    xhttp.setRequestHeader("Content-Type", "application/json");
    xhttp.setRequestHeader("Authorization", localStorage.getItem("Authorization"));
    if (window.confirm("Delete channel?")) xhttp.send();
}

async function switchChannel(event) {
    let src = event.target || event.srcElement;
    let id = src.id.replace("channel-", "");
    if (state.channel !== id) {
        return switchToChannel(id)
    }
}

async function switchToChannel(id) {
    state.channel = id
    return renderChannelHeader(state.channels[id]);
}

function renderChatbox(){
    renderChannelHeader()
}

function renderChannelHeader(channel){
    $('#channel-actions').empty()
    $('#channel-title').text('#' + channel.name)
    //$('#channel-desc').text(ch.description)

    if (state.user && channel.creator && channel.creator.UserID === state.user.id) {
        $('#channel-actions').append('<button onclick()>')
    } else {
        $('#channel-actions').empty()
    }
}

// function showChannelDetail(channelName){
//     let channelID = localStorage.getItem(channelName);

//     var xhttp = new XMLHttpRequest();
//     xhttp.onreadystatechange = function() {
//         if (this.readyState == 4) {
//             if(this.status < 300){
//                 var channels = JSON.parse(this.responseText);
//                 jQuery.each(channels, function() {
//                     if(this.id == channelID){
//                         promptSingleChannel(this);
//                     }
//                 });
//             }
//             else {
//                 handleError(this.responseText);
//             }
//         }
//     };

//     xhttp.open("GET", "https://api.garson.me:443/v1/channels", true);
//     xhttp.setRequestHeader("Content-Type", "application/json");
//     xhttp.setRequestHeader("Authorization", localStorage.getItem("Authorization"));
//     xhttp.send();
// }

function createWebsocketConnection(){
    let sock;
    let auth = localStorage.getItem("Authorization");
    sock = new WebSocket("wss://api.garson.me:443/ws?auth=" + auth);

    sock.onopen = () => {
        console.log("Connection Opened");
    };

    sock.onclose = () => {
        console.log("Connection Closed");
    };

    sock.onmessage = (msg) => {
        console.log("Message received " + msg.data);
    };
    // document.addEventListener("DOMContentLoaded", (event) => {

    // });
}