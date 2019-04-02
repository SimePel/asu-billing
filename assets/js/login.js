window.onload = function () {
    url = new URL(window.location.href);
    err = url.searchParams.get("err");
    if (err !== null) {
        document.getElementById("errorHere").innerText = "Неверный логин или пароль";
    }
};