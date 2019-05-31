fetch("users/1").then((res) => { return res.json() }).then((data) => {
    document.getElementById("name").innerHTML = data.name;
    document.getElementById("login").innerHTML = data.login;
    document.getElementById("balance").innerHTML = data.balance;
    document.getElementById("agreement").innerHTML = data.agreement;
})