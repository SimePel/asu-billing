window.onload = function () {
    let url = new URL(window.location.href);
    let type = url.searchParams.get("type");
    let name = url.searchParams.get("name");
    let account = url.searchParams.get("account");

    if (type === "wired") {
        document.getElementById("wired").innerHTML = "<strong>Проводные</strong/";
    } else if (type === "wireless") {
        document.getElementById("wireless").innerHTML = "<strong>Беспроводные</strong/";
    } else if (type === "active") {
        document.getElementById("active").innerHTML = "<strong>Включенные</strong/";
    } else if (type === "inactive") {
        document.getElementById("inactive").innerHTML = "<strong>Отключенные</strong/";
    } else if ((name === null) && (account === null)) {
        document.getElementById("all").innerHTML = "<strong>Все</strong/";
    }

    const searchButton = document.getElementById("searchButton");
    const searchInput = document.getElementById("search");

    searchButton.addEventListener('click', search);

    searchInput.addEventListener('keyup', function (event) {
        console.log(event.keyCode);
        if (event.keyCode === 13) {
            search();
        }
    });
};

function search() {
    var s = document.getElementById("select").value;
    if (s === "name") {
        window.location.replace("/adm?name=" + document.getElementById("search").value);
    } else if (s === "account") {
        window.location.replace("/adm?account=" + document.getElementById("search").value);
    } else {
        alert("Неопознанный тип поиска");
    }
}
