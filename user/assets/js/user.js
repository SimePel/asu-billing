{
    let tokenStr = document.cookie.split(";").filter((item) => item.includes("jwt"))[0].split("=")[1]
    let t = parseJWT(tokenStr)
    fetch("users/" + t.id).then((res) => { return res.json() }).then((data) => {
        document.getElementById("name").innerHTML = data.name;
        document.getElementById("login").innerHTML = data.login;
        document.getElementById("balance").innerHTML = data.balance;
        document.getElementById("agreement").innerHTML = data.agreement;
        let expiredDate = new Date(Date.parse(data.expired_date));
        document.getElementById("expired_date").innerHTML = expiredDate.getDate() + "." + (expiredDate.getMonth() + 1) + "." + expiredDate.getFullYear();

    })
}

function parseJWT(token) {
    var base64Url = token.split('.')[1];
    var base64 = decodeURIComponent(atob(base64Url).split('').map(function (c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));

    return JSON.parse(base64);
};