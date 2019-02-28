window.onload = function () {
    url = new URL(window.location.href);
    err = url.searchParams.get("err");
    if (err !== "") {
        document.getElementById("errorHere").innerText = err;
    }
};