function getUsers() {
    function createTD(tableTag, value) {
        let td = document.createElement("td");
        td.setAttribute("hidden", "");
        td.setAttribute("data-table-tag", tableTag);
        td.append(value);
        return td;
    }

    function createStatusTD(paid, activity) {
        let td = document.createElement("td");
        td.setAttribute("hidden", "");
        td.setAttribute("data-table-tag", "activity");

        let paidText = "Не оплачено. ";
        if (paid) {
            paidText = "Оплачено. ";
        }

        let span = document.createElement("span");
        span.classList.add("icon");
        span.setAttribute("title", "Без доступа в интернет");
        let i = document.createElement("i");
        i.classList.add("fas", "fa-ban");
        if (activity) {
            i.classList.replace("fa-ban", "fa-check");
            span.setAttribute("title", "Подключен к интернету");
        }
        span.append(i);

        td.append(paidText);
        td.append(span);
        return td;
    }

    function displayTable() {
        let elemsToDisplay = JSON.parse(localStorage.getItem("elemsToDisplay"));
        let defaultTable = ["name", "login", "tariff", "balance", "activity"];

        if (elemsToDisplay === null) {
            localStorage.setItem("elemsToDisplay", JSON.stringify(defaultTable));
            elemsToDisplay = defaultTable;
        }

        for (let elem of elemsToDisplay) {
            document.querySelectorAll(`[data-table-tag="${elem}"]`).forEach((td) => {
                td.removeAttribute("hidden");
            });
            document.querySelector(`[data-menu-item="${elem}"]`).classList.add("active");
        }
    }

    fetch("users")
        .then(res => {
            return res.json();
        })
        .then(users => {
            users.forEach(user => {
                let tds = [];

                tds.push(createTD("name", user.name));
                tds.push(createTD("agreement", user.agreement));
                tds.push(createTD("login", user.login));
                let expiredDate = "Не подключен";
                if (user.paid === true) {
                    const d = new Date(user.expired_date);
                    expiredDate = d.getDate() + "." + (d.getMonth() + 1) + "." + d.getFullYear();
                }
                tds.push(createTD("expiredDate", expiredDate));
                tds.push(createTD("ip", user.inner_ip));
                tds.push(createTD("phone", user.phone));
                tds.push(createTD("room", user.room));
                tds.push(createTD("tariff", user.tariff.name));
                tds.push(createTD("connectionPlace", user.connection_place));
                tds.push(createTD("balance", user.balance));
                tds.push(createStatusTD(user.paid, user.activity));

                let tr = document.createElement("tr");
                tr.append(...tds);
                tr.classList.add("clickable");
                tr.setAttribute("data-is-active", "false");
                if (user.activity) {
                    tr.setAttribute("data-is-active", "true");
                }
                tr.addEventListener("click", e => {
                    window.location.href = "/user?id=" + user.id;
                });
                document.getElementById("tbody").append(tr);
                displayTable();
            });
        });
}

function addEventListenersToMenuItems() {
    let menu = document.querySelector(".menu-list");
    menu.addEventListener("click", event => {
        let currentTable = JSON.parse(localStorage.getItem("elemsToDisplay"));
        let item = event.target;
        item.classList.toggle("active");
        let menuItemName = item.getAttribute("data-menu-item");
        if (item.classList.contains("active")) {
            document.querySelectorAll(`[data-table-tag="${menuItemName}"]`).forEach((td) => {
                td.removeAttribute("hidden");
            });
            currentTable.push(`${menuItemName}`);
        } else {
            document.querySelectorAll(`[data-table-tag="${menuItemName}"]`).forEach((td) => {
                td.setAttribute("hidden", "");
            });
            currentTable = currentTable.filter((value) => {
                return value !== menuItemName;
            });
        }
        localStorage.setItem("elemsToDisplay", JSON.stringify(currentTable));
    });
}

function showStatistics() {
    fetch("stats").then(res => {
        return res.json();
    }).then(stats => {
        document.querySelector("#countOfAllUsers").textContent = stats.active_users_count + stats.inactive_users_count;
        document.querySelector("#countOfActiveUsers").textContent = stats.active_users_count;
        document.querySelector("#countOfInactiveUsers").textContent = stats.inactive_users_count;
        document.querySelector("#allMoney").textContent = stats.all_money;
    });
}

function displayAllUsers() {
    let activeLink = document.querySelector(".active-link");
    activeLink.classList.remove("active-link");
    displayAllUsersButton.classList.add("active-link");
    removeHiddenAttributeFromAllTRs();
}

function displayOnlyActiveUsers() {
    let activeLink = document.querySelector(".active-link");
    activeLink.classList.remove("active-link");
    displayActiveUsersButton.classList.add("active-link");

    removeHiddenAttributeFromAllTRs();

    let inactiveUsers = document.querySelectorAll(`tr[data-is-active="false"]`);
    for (let tr of inactiveUsers) {
        tr.setAttribute("hidden", "");
    }
}

function displayOnlyInactiveUsers() {
    let activeLink = document.querySelector(".active-link");
    activeLink.classList.remove("active-link");
    displayInactiveUsersButton.classList.add("active-link");

    removeHiddenAttributeFromAllTRs();

    let activeUsers = document.querySelectorAll(`tr[data-is-active="true"]`);
    for (let tr of activeUsers) {
        tr.setAttribute("hidden", "");
    }
}

function removeHiddenAttributeFromAllTRs() {
    let allUsers = document.querySelectorAll(`tr`);
    for (let tr of allUsers) {
        tr.removeAttribute("hidden", "");
    }
}

getUsers();
showStatistics();
addEventListenersToMenuItems();

let displayAllUsersButton = document.querySelector("#all");
displayAllUsersButton.addEventListener("click", displayAllUsers);

let displayActiveUsersButton = document.querySelector("#active");
displayActiveUsersButton.addEventListener("click", displayOnlyActiveUsers);

let displayInactiveUsersButton = document.querySelector("#inactive");
displayInactiveUsersButton.addEventListener("click", displayOnlyInactiveUsers);

let toggle = document.querySelector(".toggle");
toggle.addEventListener("click", event => {
    toggle.classList.toggle("active");
    toggle.innerHTML == "Выкл" ?
        (toggle.innerHTML = "Вкл") :
        (toggle.innerHTML = "Выкл");
});