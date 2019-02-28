window.onload = function () {
    url = new URL(window.location.href);
    type = url.searchParams.get("type");
    if (type === "wired") {
        document.getElementById("wired").innerHTML = "<strong>Проводные</strong/";
    } else if (type === "wireless") {
        document.getElementById("wireless").innerHTML = "<strong>Беспроводные</strong/";
    } else if (type === "active") {
        document.getElementById("active").innerHTML = "<strong>Включенные</strong/";
    } else if (type === "inactive") {
        document.getElementById("inactive").innerHTML = "<strong>Отключенные</strong/";
    } else {
        document.getElementById("all").innerHTML = "<strong>Все</strong/";
    }

    const searchButton = document.getElementById("searchButton")

    searchButton.addEventListener('click', function (event) {
        window.location.href = "/admin?type=name&name=" + document.getElementById("search").value
    });
};
