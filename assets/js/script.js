function getRandomLink() {
    fetch("https://raph.ee/bookmarks.txt")
        .then( r => r.text())
        .then( function(text) {
            var lines = text.split("\n");
            var randLineNum = Math.floor(Math.random() * lines.length);
            var link = lines[randLineNum]
            document.querySelectorAll("#random-link").forEach(element => {
                element.setAttribute("href", link);
            })
        });
};
getRandomLink();
