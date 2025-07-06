async function loadCookiesAndDisplay() {
    const cookies = await chrome.cookies.getAll({ domain: ".strava.com" });
    const cookiesKeyValue = cookies.map(({ name, value }) => ({ name, value }));
    document.getElementById('cookiesOutput').textContent = JSON.stringify(cookiesKeyValue, null, 2);

    const expiresCookie = cookies.find(c => c.name === '_strava_CloudFront-Expires');
    const displayElem = document.getElementById('expiresDisplay');
    if (expiresCookie) {
        const ts = parseInt(expiresCookie.value, 10);
        const now = Date.now();
        const date = new Date(ts);
        const diffMs = ts - now;
        const diffDays = Math.round(diffMs / (1000 * 60 * 60 * 24));
        const isExpired = ts < now;
        const whenText = isExpired ? `${Math.abs(diffDays)} day${Math.abs(diffDays) !== 1 ? 's' : ''} ago` : `${diffDays} day${diffDays !== 1 ? 's' : ''}`;
        if (isExpired) {
            displayElem.textContent = `Cookies expired ${whenText} (${date.toLocaleString()})`;
            displayElem.style.color =  'red';
            exportBtn.disabled = true;
        } else {
            displayElem.textContent = `Cookies valid for ${whenText} (until ${date.toLocaleString()})`;
            displayElem.style.color = 'green';
            exportBtn.disabled = false;
        }
    } else {
        displayElem.textContent = `Some cookies are missing, are you logged in to Strava?`;
        displayElem.style.color =  'red';
        exportBtn.disabled = true;
    }
}

// on popup load
document.addEventListener("DOMContentLoaded", async () => {
    await loadCookiesAndDisplay();
    document.getElementById('exportBtn').addEventListener('click', async () => {
        const cookies = await chrome.cookies.getAll({ domain: ".strava.com" });
        const cookiesKeyValue = cookies.map(({ name, value }) => ({ name, value }));
        chrome.runtime.sendMessage({
            action: 'exportStravaCookies',
            cookies: cookiesKeyValue
        });
    });
});

// on page reload (via message from background.js)
chrome.runtime.onMessage.addListener((message) => {
    if (message.action === 'pageLoadComplete') {
        loadCookiesAndDisplay();
    }
});
