function getUsers() {
    fetch("users")
        .then(res => {
            return res.json();
        })
        .then(users => {
            fillUsersToTheTable(users);
            displayAllUsers();
            displayTable();
        });
}

function fillUsersToTheTable(users) {
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
        tr.setAttribute("data-id", user.id);
        tr.addEventListener("click", e => {
            window.location.href = "/user?id=" + user.id;
        });
        if (user.is_archived) {
            tr.setAttribute("data-is-archive", "true");
        } else {
            tr.setAttribute("data-is-active", "false");
            if (user.activity) {
                tr.setAttribute("data-is-active", "true");
            }
        }

        document.getElementById("tbody").append(tr);
    });
}

function displayTable() {
    let elemsToDisplay = JSON.parse(localStorage.getItem("elemsToDisplay"));
    let defaultTable = ["name", "login", "tariff", "balance", "activity"];

    if (elemsToDisplay === null) {
        localStorage.setItem("elemsToDisplay", JSON.stringify(defaultTable));
        elemsToDisplay = defaultTable;
    }

    for (let elem of elemsToDisplay) {
        document.querySelectorAll(`[data-table-tag="${elem}"]`).forEach(td => {
            td.removeAttribute("hidden");
        });
        document.querySelector(`[data-menu-item="${elem}"]`).classList.add("active");
    }
}

function addEventListenersToMenuItems() {
    let menu = document.querySelector(".menu-list");
    menu.addEventListener("click", event => {
        let currentTable = JSON.parse(localStorage.getItem("elemsToDisplay"));
        let item = event.target;
        item.classList.toggle("active");
        let menuItemName = item.getAttribute("data-menu-item");
        if (item.classList.contains("active")) {
            document.querySelectorAll(`[data-table-tag="${menuItemName}"]`).forEach(td => {
                td.removeAttribute("hidden");
            });
            currentTable.push(`${menuItemName}`);
        } else {
            document.querySelectorAll(`[data-table-tag="${menuItemName}"]`).forEach(td => {
                td.setAttribute("hidden", "");
            });
            currentTable = currentTable.filter(value => {
                return value !== menuItemName;
            });
        }
        localStorage.setItem("elemsToDisplay", JSON.stringify(currentTable));
    });
}

function showStatistics() {
    fetch("stats")
        .then(res => {
            return res.json();
        })
        .then(stats => {
            document.querySelector("#countOfActiveUsers").textContent = stats.active_users_count;
            document.querySelector("#countOfInactiveUsers").textContent = stats.inactive_users_count;
            document.querySelector("#cash").textContent = stats.cash + " руб.";
        });
}

function sortIfItIsNeeded() {
    let nameTH = document.querySelector('thead>tr>th[data-table-tag="name"]');
    if (nameTH.childElementCount === 2) {
        if (
            nameTH.children
                .item(1)
                .getAttribute("src")
                .includes("down")
        ) {
            sortTheTableByNameAscending();
        } else {
            sortTheTableByNameDescending();
        }
    }

    let agreementTH = document.querySelector('thead>tr>th[data-table-tag="agreement"]');
    if (agreementTH.childElementCount === 2) {
        if (
            agreementTH.children
                .item(1)
                .getAttribute("src")
                .includes("down")
        ) {
            sortTheTableByAgreementAscending();
        } else {
            sortTheTableByAgreementDescending();
        }
    }
}

function displayAllUsers() {
    let activeLink = document.querySelector(".active-link");
    if (activeLink) {
        activeLink.classList.remove("active-link");
    }
    displayAllUsersButton.classList.add("active-link");

    removeHiddenAttributeFromAll("tr");

    let archiveUsers = document.querySelectorAll(`tr[data-is-archive="true"]`);
    for (let tr of archiveUsers) {
        tr.setAttribute("hidden", "");
    }

    sortIfItIsNeeded();
}

function displayOnlyActiveUsers() {
    let activeLink = document.querySelector(".active-link");
    if (activeLink) {
        activeLink.classList.remove("active-link");
    }
    displayActiveUsersButton.classList.add("active-link");

    setHiddenAttribiteForAll("#tbody>tr");

    let activeUsers = document.querySelectorAll(`tr[data-is-active="true"]`);
    for (let tr of activeUsers) {
        tr.removeAttribute("hidden");
    }

    sortIfItIsNeeded();
}

function displayOnlyInactiveUsers() {
    let activeLink = document.querySelector(".active-link");
    if (activeLink) {
        activeLink.classList.remove("active-link");
    }
    displayInactiveUsersButton.classList.add("active-link");

    setHiddenAttribiteForAll("#tbody>tr");

    let inactiveUsers = document.querySelectorAll(`tr[data-is-active="false"]`);
    for (let tr of inactiveUsers) {
        tr.removeAttribute("hidden");
    }

    sortIfItIsNeeded();
}

function displayOnlyArchiveUsers() {
    let activeLink = document.querySelector(".active-link");
    if (activeLink) {
        activeLink.classList.remove("active-link");
    }
    displayArchiveUsersButton.classList.add("active-link");

    setHiddenAttribiteForAll("#tbody>tr");

    let archiveUsers = document.querySelectorAll(`tr[data-is-archive="true"]`);
    for (let tr of archiveUsers) {
        tr.removeAttribute("hidden");
    }

    sortIfItIsNeeded();
}

const removeHiddenAttributeFromAll = selector => {
    document.querySelectorAll(selector).forEach(elem => {
        elem.removeAttribute("hidden");
    });
};

const setHiddenAttribiteForAll = selector => {
    document.querySelectorAll(selector).forEach(elem => {
        elem.setAttribute("hidden", "");
    });
};

const searchThroughTheTable = () => {
    let whatToSearch = document.querySelector("#search").value.toLowerCase();
    let treeWalker = document.createTreeWalker(document.getElementById("tbody"), NodeFilter.SHOW_ELEMENT, {
        acceptNode: node => {
            if (node.textContent.toLowerCase().includes(whatToSearch)) {
                return NodeFilter.FILTER_ACCEPT;
            }
        },
    });

    setHiddenAttribiteForAll("#tbody>tr");
    let activeLink = document.querySelector(".active-link");
    if (activeLink) {
        activeLink.classList.remove("active-link");
    }

    for (let node = treeWalker.nextNode(); ; node = treeWalker.nextSibling()) {
        if (node == null) {
            break;
        }

        let elem = node.firstChild.parentElement;
        elem.removeAttribute("hidden");
    }
};

function getUsersFromTheTable() {
    let users = [];

    document.querySelectorAll("#tbody>tr").forEach(tr => {
        if (tr.getAttribute("hidden") === "") {
            return;
        }

        let user = {
            id: 0,
            name: "",
            agreement: "",
            login: "",
            expired_date: "Не подключен",
            inner_ip: "",
            phone: "",
            room: "",
            tariff: {
                name: "",
            },
            connection_place: "",
            balance: 0,
            is_archived: false,
            paid: false,
            activity: false,
        };
        user.id = parseInt(tr.getAttribute("data-id"));
        tr.querySelectorAll("td").forEach(td => {
            if (td.getAttribute("data-table-tag") === "name") {
                user.name = td.textContent;
            }
            if (td.getAttribute("data-table-tag") === "agreement") {
                user.agreement = td.textContent;
            }
            if (td.getAttribute("data-table-tag") === "login") {
                user.login = td.textContent;
            }
            if (td.getAttribute("data-table-tag") === "expiredDate") {
                user.expired_date = td.textContent;
            }
            if (td.getAttribute("data-table-tag") === "ip") {
                user.inner_ip = td.textContent;
            }
            if (td.getAttribute("data-table-tag") === "phone") {
                user.phone = td.textContent;
            }
            if (td.getAttribute("data-table-tag") === "room") {
                user.room = td.textContent;
            }
            if (td.getAttribute("data-table-tag") === "tariff") {
                user.tariff.name = td.textContent;
            }
            if (td.getAttribute("data-table-tag") === "connectionPlace") {
                user.connection_place = td.textContent;
            }
            if (td.getAttribute("data-table-tag") === "balance") {
                user.balance = parseInt(td.textContent);
            }
            if (td.getAttribute("data-table-tag") === "activity") {
                if (td.textContent === "Оплачено.") {
                    user.paid = true;
                } else if (td.textContent === "В архиве") {
                    user.is_archived = true;
                }
                let spanClassList = td.lastChild.classList;
                if (spanClassList !== undefined) {
                    if (spanClassList.contains("has-text-success")) {
                        user.activity = true;
                    }
                }
            }
        });
        users.push(user);
    });

    return users;
}

function removeOnlyVisibleTRs() {
    document.querySelectorAll("#tbody>tr").forEach(tr => {
        if (tr.getAttribute("hidden") !== "") {
            tr.remove();
        }
    });
}

function removeAllImgsInTheTHs() {
    let imgNodes = document.querySelectorAll("thead>tr>th>img");
    if (imgNodes !== null) {
        imgNodes.forEach(img => {
            img.remove();
        });
    }
}

function sortTheTableByNameAscending() {
    let users = getUsersFromTheTable();
    removeOnlyVisibleTRs();
    removeAllImgsInTheTHs();

    let arrow = document.createElement("img");
    arrow.src = "../assets/img/icons8-down-arrow-20.png";
    document.querySelector('thead>tr>th[data-table-tag="name"]').appendChild(arrow);

    users.sort(function(a, b) {
        var nameA = a.name;
        var nameB = b.name;
        if (nameA < nameB) {
            return -1;
        }
        if (nameA > nameB) {
            return 1;
        }
        return 0;
    });
    fillUsersToTheTable(users);
    displayTable();

    let nameButton = document.querySelector('thead>tr>th[data-table-tag="name"]>a');
    nameButton.removeEventListener("click", sortTheTableByNameAscending);
    nameButton.addEventListener("click", sortTheTableByNameDescending);
}

function sortTheTableByNameDescending() {
    let users = getUsersFromTheTable();
    removeOnlyVisibleTRs();
    removeAllImgsInTheTHs();

    let arrow = document.createElement("img");
    arrow.src = "../assets/img/icons8-up-20.png";
    document.querySelector('thead>tr>th[data-table-tag="name"]').appendChild(arrow);

    users.sort(function(a, b) {
        var nameA = a.name;
        var nameB = b.name;
        if (nameA < nameB) {
            return 1;
        }
        if (nameA > nameB) {
            return -1;
        }
        return 0;
    });
    fillUsersToTheTable(users);
    displayTable();

    let nameButton = document.querySelector('thead>tr>th[data-table-tag="name"]>a');
    nameButton.removeEventListener("click", sortTheTableByNameDescending);
    nameButton.addEventListener("click", sortTheTableByNameAscending);
}

function sortTheTableByAgreementAscending() {
    let users = getUsersFromTheTable();
    removeOnlyVisibleTRs();
    removeAllImgsInTheTHs();

    let arrow = document.createElement("img");
    arrow.src = "../assets/img/icons8-down-arrow-20.png";
    document.querySelector('thead>tr>th[data-table-tag="agreement"]').appendChild(arrow);

    users.sort(function(a, b) {
        var agreementA = a.agreement;
        var agreementB = b.agreement;
        if (agreementA < agreementB) {
            return -1;
        }
        if (agreementA > agreementB) {
            return 1;
        }
        return 0;
    });
    fillUsersToTheTable(users);
    displayTable();

    let agreementButton = document.querySelector('thead>tr>th[data-table-tag="agreement"]>a');
    agreementButton.removeEventListener("click", sortTheTableByAgreementAscending);
    agreementButton.addEventListener("click", sortTheTableByAgreementDescending);
}

function sortTheTableByAgreementDescending() {
    let users = getUsersFromTheTable();
    removeOnlyVisibleTRs();
    removeAllImgsInTheTHs();

    let arrow = document.createElement("img");
    arrow.src = "../assets/img/icons8-up-20.png";
    document.querySelector('thead>tr>th[data-table-tag="agreement"]').appendChild(arrow);

    users.sort(function(a, b) {
        var agreementA = a.agreement;
        var agreementB = b.agreement;
        if (agreementA < agreementB) {
            return 1;
        }
        if (agreementA > agreementB) {
            return -1;
        }
        return 0;
    });
    fillUsersToTheTable(users);
    displayTable();

    let agreementButton = document.querySelector('thead>tr>th[data-table-tag="agreement"]>a');
    agreementButton.removeEventListener("click", sortTheTableByAgreementDescending);
    agreementButton.addEventListener("click", sortTheTableByAgreementAscending);
}

function getNotificationStatus() {
    fetch("notification-status")
        .then(res => res.text())
        .then(status => {
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

let displayAllUsersButton = document.querySelector("#all");
displayAllUsersButton.addEventListener("click", displayAllUsers);

getUsers();
showStatistics();
addEventListenersToMenuItems();
getNotificationStatus();

let displayActiveUsersButton = document.querySelector("#active");
displayActiveUsersButton.addEventListener("click", displayOnlyActiveUsers);

let displayInactiveUsersButton = document.querySelector("#inactive");
displayInactiveUsersButton.addEventListener("click", displayOnlyInactiveUsers);

let displayArchiveUsersButton = document.querySelector("#archive");
displayArchiveUsersButton.addEventListener("click", displayOnlyArchiveUsers);

let nameButton = document.querySelector('thead>tr>th[data-table-tag="name"]>a');
nameButton.addEventListener("click", sortTheTableByNameAscending);

let agreementButton = document.querySelector('thead>tr>th[data-table-tag="agreement"]>a');
agreementButton.addEventListener("click", sortTheTableByAgreementAscending);

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
    smsStatus.textContent == "Включены" ? (smsStatus.textContent = "Выключены") : (smsStatus.textContent = "Включены");

    changeNotificationStatus(newStatus);
});
