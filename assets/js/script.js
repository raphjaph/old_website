function getRandomLink() {
    fetch("https://raph.8el.eu/bookmarks.txt")
        .then( r => r.text())
        .then( function(text) {
            var lines = text.split("\n");
            var randLineNum = Math.floor(Math.random() * lines.length);
            var link = lines[randLineNum]
            document.getElementById("random-link").setAttribute("href", link)
        });
};
getRandomLink();
