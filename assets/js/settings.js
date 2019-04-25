window.onload = () => {
    const emailBtn = document.getElementById("sendEmail");

    emailBtn.addEventListener('click', (event) => {
        const emailValue = document.getElementById("email").value;

        if (emailValue === "") {
            alert("Заполните поле email");
            return;
        }

        let email = { "email": emailValue };
        fetch('/settings', {
            method: "POST",
            mode: "no-cors",
            cache: "no-cache",
            credentials: "same-origin",
            headers: {
                "Content-Type": "application/json",
            },
            redirect: "follow",
            body: JSON.stringify(email),
        });

        document.getElementById("emailBox").remove();
        document.getElementById("emailNotification").classList.remove("invisible");
    });
}