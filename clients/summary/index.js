(function(){
    "use strict";
    window.onload = function() {
        document.getElementById("submit").onclick = getSummary;
    }
    function getSummary() {
        var ajax = new XMLHttpRequest();
        var input = document.getElementById("entry");
        ajax.open("GET", "http://localhost:4000/v1/summary/?url=" + input.value, true);
        ajax.onreadystatechange = function() {
            if (this.readyState == 4 && this.status == 200) {
                var json = JSON.parse(this.responseText);
                console.log(json);

                var results = document.getElementById("results");
                results.innerHTML = "";

                // title
                if (json.hasOwnProperty('title')) {
                    var title = document.createElement("h1");
                    title.innerHTML = json.title;
                    results.appendChild(title);
                }
                

                // description
                if (json.hasOwnProperty('description')) {
                    var desc = document.createElement("h3");
                    desc.innerHTML = json.description;
                    results.appendChild(desc);
                }
                
                // images
                if (json.hasOwnProperty('images')) {
                    var images = document.createElement("div");
                    for (var i in json.images) {
                        var img = document.createElement("img");
                        img.src = json.images[i].url;
                        images.appendChild(img);
                    }
                    results.appendChild(images);
                }
            }
        }
        ajax.send();
    }
})();