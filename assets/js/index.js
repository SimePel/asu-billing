function getUsers() {
    function createTD(tableTag, value) {
        let td = document.createElement("td");
        td.setAttribute("hidden", "");
        td.setAttribute("data-table-tag", tableTag);
        td.append(value);
        return td;
    }

    function createStatusTD(paid, activity, is_archived) {
        let td = document.createElement("td");
        td.setAttribute("hidden", "");
        td.setAttribute("data-table-tag", "activity");
        if (is_archived) {
            td.append("В архиве");
            return td;
        }

        let paidText = "Не оплачено. ";
        if (paid) {
            paidText = "Оплачено. ";
        }

        let span = document.createElement("span");
        span.classList.add("icon", "has-text-danger");
        span.setAttribute("title", "Без доступа в интернет");
        let i = document.createElement("i");
        i.classList.add("fas", "fa-ban");
        if (activity) {
            i.classList.replace("fa-ban", "fa-check");
            span.classList.replace("has-text-danger", "has-text-success");
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
                tds.push(createStatusTD(user.paid, user.activity, user.is_archived));

                let tr = document.createElement("tr");
                tr.append(...tds);
                tr.classList.add("clickable");
                tr.addEventListener("click", e => {
                    window.location.href = "/user?id=" + user.id;
                });
                if (user.is_archived) {
                    tr.setAttribute("data-is-archive", "true");
                    tr.setAttribute("hidden", "");
                } else {
                    tr.setAttribute("data-is-active", "false");
                    if (user.activity) {
                        tr.setAttribute("data-is-active", "true");
                    }
                }

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
        document.querySelector("#countOfActiveUsers").textContent = stats.active_users_count;
        document.querySelector("#countOfInactiveUsers").textContent = stats.inactive_users_count;
    });
}

function displayAllUsers() {
    let activeLink = document.querySelector(".active-link");
    if (activeLink) {
        activeLink.classList.remove("active-link");
    }
    displayAllUsersButton.classList.add("active-link");

    removeHiddenAttributeForAllTRs();

    let archiveUsers = document.querySelectorAll(`tr[data-is-archive="true"]`);
    for (let tr of archiveUsers) {
        tr.setAttribute("hidden", "");
    }
}

function displayOnlyActiveUsers() {
    let activeLink = document.querySelector(".active-link");
    if (activeLink) {
        activeLink.classList.remove("active-link");
    }
    displayActiveUsersButton.classList.add("active-link");

    setHiddenAttribiteForAllTRs();

    let activeUsers = document.querySelectorAll(`tr[data-is-active="true"]`);
    for (let tr of activeUsers) {
        tr.removeAttribute("hidden");
    }
}

function displayOnlyInactiveUsers() {
    let activeLink = document.querySelector(".active-link");
    if (activeLink) {
        activeLink.classList.remove("active-link");
    }
    displayInactiveUsersButton.classList.add("active-link");

    setHiddenAttribiteForAllTRs();

    let inactiveUsers = document.querySelectorAll(`tr[data-is-active="false"]`);
    for (let tr of inactiveUsers) {
        tr.removeAttribute("hidden");
    }
}

function displayOnlyArchiveUsers() {
    let activeLink = document.querySelector(".active-link");
    if (activeLink) {
        activeLink.classList.remove("active-link");
    }
    displayArchiveUsersButton.classList.add("active-link");

    setHiddenAttribiteForAllTRs();

    let archiveUsers = document.querySelectorAll(`tr[data-is-archive="true"]`);
    for (let tr of archiveUsers) {
        tr.removeAttribute("hidden");
    }
}

function removeHiddenAttributeForAllTRs() {
    let allTRs = document.querySelectorAll(`tr`);
    for (let tr of allTRs) {
        tr.removeAttribute("hidden");
    }
}

function setHiddenAttribiteForAllTRs() {
    let allTRs = document.querySelectorAll(`tbody>tr`);
    for (let tr of allTRs) {
        tr.setAttribute("hidden", "");
    }
}

function searchThroughTheTable() {
    let whatToSearch = document.querySelector("#search").value;
    whatToSearch = whatToSearch.toLowerCase();
    let treeWalker = document.createTreeWalker(
        document.getElementById("tbody"),
        NodeFilter.SHOW_ELEMENT, {
            acceptNode: function (node) {
                if (node.textContent.toLowerCase().includes(whatToSearch)) {
                    return NodeFilter.FILTER_ACCEPT;
                }
            }
        }
    );

    setHiddenAttribiteForAllTRs();
    let activeLink = document.querySelector(".active-link");
    if (activeLink) {
        activeLink.classList.remove("active-link");
    }

    for (let node = treeWalker.nextNode();; node = treeWalker.nextSibling()) {
        if (node == null) {
            break;
        }

        let elem = node.firstChild.parentElement;
        elem.removeAttribute("hidden");
    }
}

function getNotificationStatus() {
    fetch("notification-status").then(res => res.text()).then(status => {
        if (status === "false") {
            smsStatus.classList.add("disable");
            smsStatus.textContent = "Выключены";
        }
    });
}

function changeNotificationStatus(newStatus) {
    fetch("change-notification-status", {
        method: "POST",
        body: newStatus,
    });
}

getUsers();
showStatistics();
addEventListenersToMenuItems();
getNotificationStatus();

let displayAllUsersButton = document.querySelector("#all");
displayAllUsersButton.addEventListener("click", displayAllUsers);

let displayActiveUsersButton = document.querySelector("#active");
displayActiveUsersButton.addEventListener("click", displayOnlyActiveUsers);

let displayInactiveUsersButton = document.querySelector("#inactive");
displayInactiveUsersButton.addEventListener("click", displayOnlyInactiveUsers);

let displayArchiveUsersButton = document.querySelector("#archive");
displayArchiveUsersButton.addEventListener("click", displayOnlyArchiveUsers);

let searchButton = document.querySelector("#searchButton");
searchButton.addEventListener("click", searchThroughTheTable);
let searchInput = document.querySelector("#search");
searchInput.addEventListener("keyup", event => {
    if (event.keyCode === 13) {
        searchThroughTheTable();
    }
});

let smsStatus = document.querySelector("#sms-status");
smsStatus.addEventListener("click", () => {
    let newStatus = !smsStatus.classList.toggle("disable");
    smsStatus.textContent == "Включены" ?
        (smsStatus.textContent = "Выключены") :
        (smsStatus.textContent = "Включены");

    changeNotificationStatus(newStatus);
});