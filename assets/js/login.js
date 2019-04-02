window.onload = function () {
    url = new URL(window.location.href);
    err = url.searchParams.get("err");
    if (err == "1") {
        document.getElementById("errorHere").innerText = "Неверный логин или пароль";
    } else if (err == "2") {
        document.getElementById("errorHere").innerText = "Вы не подключены к биллинговой системе";
    } else {
        ;
    }
};