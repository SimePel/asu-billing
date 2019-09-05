function getUsers() {
    function createTD(value) {
        let td = document.createElement("td");
        td.append(value);
        return td;
    }

    function createStatusIcon(activity) {
        let td = document.createElement("td");
        let span = document.createElement("span");
        let i = document.createElement("i");
        if (activity === true) {
            span.classList.add("icon", "has-text-success");
            i.classList.add("fas", "fa-user-check");
        } else {
            span.classList.add("icon", "has-text-danger");
            i.classList.add("fas", "fa-user-times");
        }
        span.append(i);
        td.append(span);

        return td;
    }

    fetch("users")
        .then(res => {
            return res.json();
        })
        .then(users => {
            users.forEach(user => {
                let tds = [];

                tds.push(createTD(user.name));
                tds.push(createTD(user.agreement));
                tds.push(createTD(user.login));
                let expiredDate = "Не подключен";
                if (user.activity === true) {
                    const d = new Date(user.expired_date);
                    expiredDate = d.getDay() + "." + d.getMonth() + "." + d.getFullYear();
                }
                tds.push(createTD(expiredDate));
                tds.push(createTD(user.inner_ip));
                tds.push(createTD(user.phone));
                tds.push(createTD(user.room));
                tds.push(createTD(user.tariff.name));
                tds.push(createTD(user.connection_place));
                tds.push(createTD(user.balance));
                tds.push(createStatusIcon(user.activity));

                let tr = document.createElement("tr");
                tr.append(...tds);
                tr.classList.add("clickable");
                tr.addEventListener("click", e => {
                    window.location.href = "/user?id=" + user.id;
                });
                document.getElementById("tbody").append(tr);
            });
        });
}

getUsers();

let menu = document.querySelector(".menu-list");
menu.addEventListener("click", event => {
    let item = event.target;
    item.classList.toggle("active");
});

let toggle = document.querySelector(".toggle");
toggle.addEventListener("click", event => {
    toggle.classList.toggle("active");
    toggle.innerHTML == "Выкл"
        ? (toggle.innerHTML = "Вкл")
        : (toggle.innerHTML = "Выкл");
});
