function getUsers() {
    function createTD(value) {
        let td = document.createElement("td");
        td.append(value);
        return td;
    }

    fetch("users").then((res) => { return res.json() }).then((users) => {
        users.forEach(user => {
            let tds = [];

            tds.push(createTD(user.name));
            tds.push(createTD(user.agreement));
            tds.push(createTD(user.login));
            let d = new Date(user.expired_date);
            let expiredDate = "Не подключен";
            if (d.getFullYear() !== 1) {
                expiredDate = d.getDay() + "." + d.getMonth() + "." + d.getFullYear();
            }
            tds.push(createTD(expiredDate));
            tds.push(createTD(user.inner_ip));
            tds.push(createTD(user.phone));
            tds.push(createTD(user.room));
            tds.push(createTD(user.tariff.name));
            tds.push(createTD(user.connection_place));
            tds.push(createTD(user.balance));
            tds.push(createTD(user.activity));

            let tr = document.createElement("tr");
            tr.append(...tds);
            document.getElementById("tbody").append(tr);
        });
    })
}

window.onload = () => {
    getUsers();
}