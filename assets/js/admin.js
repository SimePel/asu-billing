window.onload = function () {
    const searchButton = document.getElementById("searchButton")

    searchButton.addEventListener('click', function (event) {
        window.location.href = "/admin?type=name&name=" + document.getElementById("search").value
    });
};
