function auth() {
    fetch("login", {
        method: "POST",
        headers: { "Content-Type": "application/json; charset=utf-8" },
        body: JSON.stringify({
            login: document.getElementById("login").value,
            password: document.getElementById("password").value,
        }),
    }).then((res) => { return res.json() }).then((data) => {
        if (data.answer != "ok") {
            document.getElementById("errorHere").innerHTML = data.error;
            return;
        }

        window.location.replace("http://localhost:8081/");
    })
}

document.getElementById("loginBtn").addEventListener("click", auth)
document.getElementById("password").addEventListener("keyup", (event) => {
    if (event.keyCode === 13) {
        auth();
    }
})