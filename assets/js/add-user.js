async function fillAgreementField() {
    let response = await fetch("next-agreement");
    if (!response.ok) {
        alert("Ошибка HTTP: " + response.status);
        return;
    }

    let json = await response.json();
    document.getElementsByName("agreement")[0].value = json.agreement;
}

fillAgreementField();