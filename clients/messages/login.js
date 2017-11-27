//@ts-check
"use strict";

//All errors returned from your APIs must be shown to the user. 
//You can write the error to the console, but you must also show it to the user via some sort of UI element. Using alert() is OK, but not elegant.
//The user must be able to sign-up for a new account, or sign-in to an existing account.
//The sign-in form must allow the user to enter the fields required by the users.Credentials struct: email address and password.

//When the user attempts to sign-up or sign-in, prevent the browser's default form-submit behavior
//and use AJAX to POST the new user or credentials data as a JSON object to the appropriate API (/v1/users for sign-up /v1/sessions for sign-in.)
//If you get a successful response (status code < 300), save the contents of the Authorization response header to local storage.
//Send this value in the Authorization request header with every subsequent AJAX request you send to your API server.

//After successfully authenticating, the user must be shown a view that displays the user's name, and allows the user to sign-out.
//This will eventually become the main view where users read messages posted by others, and post new messages.
//The user must not be able to see this view if the user is not authenticated.
//If the user is not authenticated, the user must be taken back to the sign-in view.

//When the user signs-out, use AJAX to send a DELETE request to /v1/sessions/mine.
//If you get a successful response, delete the Authorization token you previously stored in local storage, and return to the sign-in view.

//This view must also provide some UI element that takes the user to another view to update the user profile.
//This view should let the user update the first and last name fields only. Use AJAX to send a PATCH request to your /v1/users/me API, including those updates a JSON object in the request body.
//After successfully updating, return the user to the main view.

window.onload = function() {
    //sessions and users
    document.getElementById("channel-update").style.display = "none";
    document.getElementById("message-update").style.display = "none";
    document.getElementById("message-post").style.display = "none";
    document.getElementById("edit-view").style.display = "none";
    document.getElementById("edit-button").onclick = (function(){
        document.getElementById("main-view").style.display = "none";
        document.getElementById("edit-view").style.display = "block";
    });

    if (localStorage.getItem("auth") != null) {
        document.getElementById("login-view").style.display = "none";
        document.getElementById("main-view").style.display = "block";
        loadName();
    } else {
        document.getElementById("login-view").style.display = "block";
        document.getElementById("main-view").style.display = "none";
    }

    document.getElementById("sign-up").onsubmit = (function(e) {
        e.preventDefault();
        var ajax = new XMLHttpRequest();

        var email = document.getElementById("su-email").value;
        var username = document.getElementById("su-username").value;
        var fname = document.getElementById("su-fname").value;
        var lname = document.getElementById("su-lname").value;
        var password = document.getElementById("su-password").value;
        var passwordconf = document.getElementById("su-passwordconf").value;

        var obj = {
            "Email": email,
            "UserName": username,
            "FirstName": fname,
            "LastName": lname,
            "Password": password,
            "PasswordConf": passwordconf
        };

        var json = JSON.stringify(obj);

        ajax.open("POST", "https://api.fredhw.me/v1/users/", true);
        ajax.onreadystatechange = function() {
            if (this.readyState == 4 && this.status < 300) {
                var auth = this.getResponseHeader("Authorization");
                console.log(auth);
                localStorage.setItem('auth', auth);
                location.reload();
            } else if (this.readyState == 4 && this.status >= 300) {
                var err = document.getElementById("err");
                err.innerHTML = "error: " + this.responseText;
            }
        }
        ajax.send(json);
    });

    document.getElementById("sign-in").onsubmit = (function(e) {
        e.preventDefault();
        var ajax = new XMLHttpRequest();

        var email = document.getElementById("si-email").value;
        var password = document.getElementById("si-password").value;

        var obj = {
            "Email": email,
            "Password": password
        };

        var json = JSON.stringify(obj);

        ajax.open("POST", "https://api.fredhw.me/v1/sessions/", true);
        ajax.onreadystatechange = function() {
            if (this.readyState == 4 && this.status < 300) {
                var auth = this.getResponseHeader("Authorization");
                console.log(auth);
                localStorage.setItem('auth', auth);
                location.reload();
            } else if (this.readyState == 4 && this.status >= 300) {
                var err = document.getElementById("err");
                err.innerHTML = "error: " + this.responseText;
            }
        }
        ajax.send(json);
    });

    document.getElementById("edit-name").onsubmit = (function(e) {
        e.preventDefault();
        var ajax = new XMLHttpRequest();

        var auth = localStorage.getItem("auth");
        var fname = document.getElementById("en-fname").value;
        var lname = document.getElementById("en-lname").value;

        var obj = {
            "FirstName": fname,
            "LastName": lname
        };

        var json = JSON.stringify(obj);

        ajax.open("PATCH", "https://api.fredhw.me/v1/users/me", true);
        ajax.setRequestHeader("Authorization", auth);
        ajax.onreadystatechange = function() {
            if (this.readyState == 4 && this.status < 300) {
                location.reload();
            } else if (this.readyState == 4 && this.status >= 300) {
                var err = document.getElementById("err");
                err.innerHTML = "error: " + this.responseText;
            }
        }
        ajax.send(json);
    });

    document.getElementById("sbar").addEventListener('input', doSearch, false);

    document.getElementById("logout").onclick = (function() {
        var ajax = new XMLHttpRequest();

        var auth = localStorage.getItem("auth");

        ajax.open("DELETE", "https://api.fredhw.me/v1/sessions/mine/", true);
        ajax.setRequestHeader("Authorization", auth);
        ajax.onreadystatechange = function() {
            if (this.readyState == 4 && this.status < 300) {
                console.log(this.responseText);
                localStorage.removeItem("auth");
                location.reload();
            } else if (this.readyState == 4 && this.status >= 300) {
                var err = document.getElementById("err");
                err.innerHTML = this.responseText;
            }
        }
        ajax.send();
    });
}

function doSearch() {
    console.log("keyup");
    var ajax = new XMLHttpRequest();
    var sbar = document.getElementById("sbar");
    var auth = localStorage.getItem("auth");
    ajax.open("GET", "https://api.fredhw.me/v1/users?q=" + sbar.value, true);
    ajax.setRequestHeader("Authorization", auth);
    ajax.onreadystatechange = function() {
        if (this.readyState == 4 && this.status < 300) {
            var json = JSON.parse(this.responseText);
            console.log(json);
            
            var dropdown = document.getElementById("dropdown");
            dropdown.innerHTML = "";
            for (var i in json) {
                var opt = document.createElement("div");
                var username = document.createElement("h3");
                username.innerHTML = json[i].userName;
                var name = document.createElement("p");
                name.innerHTML = json[i].firstName + " " + json[i].lastName;
                var email = document.createElement("p");
                email.innerHTML = json[i].email;
                opt.appendChild(username);
                opt.appendChild(name);
                opt.appendChild(email);
                opt.classList.add("opt");
                dropdown.appendChild(opt);
            }

        } else if (this.readyState == 4 && this.status >= 300) {
            var err = document.getElementById("err");
            err.innerHTML = this.responseText;
        }
    }
    ajax.send();
}

function loadName() {
    var ajax = new XMLHttpRequest();
    
    var auth = localStorage.getItem("auth");

    ajax.open("GET", "https://api.fredhw.me/v1/users/me", true);
    ajax.setRequestHeader("Authorization", auth);
    ajax.onreadystatechange = function() {
        if (this.readyState == 4 && this.status < 300) {
            var result = JSON.parse(this.responseText);
            console.log(result);
            var field = document.getElementById("mv-name");
            field.innerHTML = result["firstName"] + " " + result["lastName"];
        } else if (this.readyState == 4 && this.status >= 300) {
            var err = document.getElementById("err");
            err.innerHTML = "error: " + this.responseText;
        }
    }
    ajax.send();


    //websocket api
    const status = document.querySelector("#status");
    const notifications = document.querySelector("#notifications");
    const errors = document.querySelector("#errors");
    
    const websocket = new WebSocket("wss://api.fredhw.me/v1/ws?auth=" + auth);
    websocket.addEventListener("error", function(err) {
        console.log(err.message);
    });
    websocket.addEventListener("open", function() {
        console.log("open");
    });
    websocket.addEventListener("close", function() {
        console.log("closed");
    });
    websocket.addEventListener("message", function(event) {
        console.log(event.data);
    });

    document.querySelector("#channel-new").addEventListener("submit", function(e) {
        e.preventDefault();
        var ajax = new XMLHttpRequest();

        var name = document.getElementById("ch-new-name").value;
        var desc = document.getElementById("ch-new-desc").value;

        var obj = {
            "name": name,
            "description": desc
        };

        var json = JSON.stringify(obj);
        
        ajax.open("POST", "https://api.fredhw.me/v1/channels/", true);
        ajax.setRequestHeader("Authorization", auth);
        ajax.onreadystatechange = function() {
            if (this.readyState == 4 && this.status < 300) {
                loadChannels();
            } else if (this.readyState == 4 && this.status >= 300) {
                var err = document.getElementById("err");
                err.innerHTML = "error: " + this.responseText;
            }
        }
        ajax.send(json);
    });

    loadChannels();

    document.querySelector("#channel-update").addEventListener("submit", function(e) {
        e.preventDefault();
        var ajax = new XMLHttpRequest();

        var name = document.getElementById("ch-upd-name").value;
        var desc = document.getElementById("ch-upd-desc").value;

        var obj = {
            "name": name,
            "description": desc
        };

        var json = JSON.stringify(obj);
        var chid = document.getElementById("ch-id").value;
        
        ajax.open("PATCH", "https://api.fredhw.me/v1/channels/" + chid, true);
        ajax.setRequestHeader("Authorization", auth);
        ajax.onreadystatechange = function() {
            if (this.readyState == 4 && this.status < 300) {
                document.getElementById("channel-update").style.display = "none";
                loadChannels();
            } else if (this.readyState == 4 && this.status >= 300) {
                var err = document.getElementById("err");
                err.innerHTML = "error: " + this.responseText;
            }
        }
        ajax.send(json);
    });    

    document.querySelector("#message-update").addEventListener("submit", function(e) {
        e.preventDefault();
        var ajax = new XMLHttpRequest();

        var body = document.getElementById("ms-upd-body").value;

        var obj = {
            "body": body
        };

        var json = JSON.stringify(obj);
        var msid = document.getElementById("ms-id").value;
        
        ajax.open("PATCH", "https://api.fredhw.me/v1/messages/" + msid, true);
        ajax.setRequestHeader("Authorization", auth);
        ajax.onreadystatechange = function() {
            if (this.readyState == 4 && this.status < 300) {
                document.getElementById("message-update").style.display = "none";
                loadMessages();
            } else if (this.readyState == 4 && this.status >= 300) {
                var err = document.getElementById("err");
                err.innerHTML = "error: " + this.responseText;
            }
        }
        ajax.send(json);
    }); 

    document.querySelector("#message-post").addEventListener("submit", function(e) {
        e.preventDefault();
        var ajax = new XMLHttpRequest();

        var body = document.getElementById("ms-post-body").value;

        var obj = {
            "body": body
        };

        var json = JSON.stringify(obj);
        var chid = document.getElementById("ch-id").value;
        
        ajax.open("POST", "https://api.fredhw.me/v1/channels/" + chid, true);
        ajax.setRequestHeader("Authorization", auth);
        ajax.onreadystatechange = function() {
            if (this.readyState == 4 && this.status < 300) {
                loadMessages();
            } else if (this.readyState == 4 && this.status >= 300) {
                var err = document.getElementById("err");
                err.innerHTML = "error: " + this.responseText;
            }
        }
        ajax.send(json);
    }); 
}

function loadChannels() {
    var ajax = new XMLHttpRequest();
    var auth = localStorage.getItem("auth");
    
    var list = document.querySelector("#channel-list");
    list.innerHTML = "";
    
    ajax.open("GET", "https://api.fredhw.me/v1/channels/", true);
    ajax.setRequestHeader("Authorization", auth);
    ajax.onreadystatechange = function() {
        if (this.readyState == 4 && this.status < 300) {
            var result = JSON.parse(this.responseText);
            console.log(result);
            for (var i in result) {
                var container = document.createElement("li");
                var link = document.createElement("a");
                link.innerHTML = result[i].name;
                link.addEventListener("click", function() {
                    var hiddench = document.getElementById("ch-id");
                    hiddench.value = result[i]["_id"];
                    console.log(hiddench.value);
                    loadMessages();
                });

                var upd = document.createElement("button");
                upd.innerHTML = "update";
                upd.addEventListener("click", function() {
                    document.getElementById("channel-update").style.display = "block";
                });
                
                var del = document.createElement("button");
                del.innerHTML = "delete";
                del.addEventListener("click", function() {
                    var ajax = new XMLHttpRequest();
                    
                    ajax.open("DELETE", "https://api.fredhw.me/v1/channels/" + result[i]._id , true);
                    ajax.setRequestHeader("Authorization", auth);
                    ajax.onreadystatechange = function() {
                        if (this.readyState == 4 && this.status < 300) {
                            document.getElementById("channel-list").innerHTML = "";
                            document.getElementById("message-list").innerHTML = "";
                            document.getElementById("channel-update").style.display = "none";
                            document.getElementById("message-update").style.display = "none";
                            document.getElementById("message-post").style.display = "none";
                            loadChannels();
                        } else if (this.readyState == 4 && this.status >= 300) {
                            var err = document.getElementById("err");
                            err.innerHTML = "error: " + this.responseText;
                        }
                    }
                    ajax.send();
                });
                container.appendChild(link);
                container.appendChild(upd);
                container.appendChild(del);
                list.appendChild(container);
            }
        } else if (this.readyState == 4 && this.status >= 300) {
            var err = document.getElementById("err");
            err.innerHTML = "error: " + this.responseText;
        }
    }
    ajax.send();
}

function loadMessages() {
    var ajax = new XMLHttpRequest();
    var auth = localStorage.getItem("auth");
    var chid = document.getElementById("ch-id").value;
    var list = document.getElementById("message-list");
    document.getElementById("message-post").style.display = "block";
    list.innerHTML = "";

    ajax.open("GET", "https://api.fredhw.me/v1/channels/" + chid, true);
    ajax.setRequestHeader("Authorization", auth);
    ajax.onreadystatechange = function() {
        if (this.readyState == 4 && this.status < 300) {
            var result = JSON.parse(this.responseText);
            console.log(result);
            for (var i in result) {
                var container = document.createElement("li");
                var link = document.createElement("p");
                link.innerHTML = result[i].name;

                var upd = document.createElement("button");
                upd.innerHTML = "update";
                upd.addEventListener("click", function() {
                    var msid = document.getElementById("ms-id");
                    msid.value = result[i]["_id"];
                    console.log(msid.value);
                    document.getElementById("message-update").style.display = "block";
                });
                
                var del = document.createElement("button");
                del.innerHTML = "delete";
                del.addEventListener("click", function() {
                    var ajax = new XMLHttpRequest();
                    
                    ajax.open("DELETE", "https://api.fredhw.me/v1/messages/" + result[i]._id , true);
                    ajax.setRequestHeader("Authorization", auth);
                    ajax.onreadystatechange = function() {
                        if (this.readyState == 4 && this.status < 300) {
    
                            document.getElementById("message-update").style.display = "none";
                            
                            loadMessages();
                        } else if (this.readyState == 4 && this.status >= 300) {
                            var err = document.getElementById("err");
                            err.innerHTML = "error: " + this.responseText;
                        }
                    }
                    ajax.send();
                });
                container.appendChild(link);
                container.appendChild(upd);
                container.appendChild(del);
                list.appendChild(container);
            }
        } else if (this.readyState == 4 && this.status >= 300) {
            var err = document.getElementById("err");
            err.innerHTML = "error: " + this.responseText;
        }
    }
    ajax.send();
}