// severity: ok, error
const setWsLocalStatusText = (severity, text) => {   
    const elText = document.getElementById("ws-local-status")
    elText.innerHTML = text;

    const elHeader = document.getElementById("bj-navbar")
    if (elHeader) {
        elHeader.classList.remove("bg-danger", "bg-success");
        switch(severity) {
            case 'ok':
                elHeader.classList.add("bg-success");
                break;
            case 'error':
                elHeader.classList.add("bg-danger");
                break;
        }
    }
}

// severity: ok, error
const setWsRemoteStatusText = (severity, text) => {
    const el = document.getElementById("ws-remote-status")
    el.innerHTML = text;
    el.className = severity == 'ok' ? 'text-success' : 'text-danger';
}

const getDataFromForm = (elForm) => {
    return Object.fromEntries(new FormData(elForm))
}

// #TODO: catch errors
class BjmrWebSocket {
    conn = null;

    constructor() {
    }

    open = (url) => {
        if (!('WebSocket' in window)) {
            console.error('WebSocket is not supported by your browser.');
            return null;
        }

        this.conn = new WebSocket(url);
        console.log(url);
        console.log(this.conn);
        return this.conn;
    }

    // #TODO: make better!
    isOpen = () => {
        return !!this.conn;
    }

    sendMsg = (type, data) => {
        if (!this.isOpen()) {
            return false;
        }

        const body = {
            type: type,
            data: data
        };

        const blob = new Blob([JSON.stringify(body, null, 2)], {
            type: "application/json",
        });

        this.conn.send(blob);
    }

}

class BjmrAjax {
    sessionKey = null;

    constructor(sessionKey) {
        this.sessionKey = sessionKey;
    }

    makeReq = async(method, endpoint, data) => {
        const out = {
            ok: false,
            code: null,
            data: null,
        }

        try {
            const response = await fetch(`/${this.sessionKey}/api/${endpoint}.json`, {
                method: method,
                // mode: "cors", // no-cors, *cors, same-origin
                cache: "no-cache", // *default, no-cache, reload, force-cache, only-if-cached
            // credentials: "same-origin", // include, *same-origin, omit
                headers: {
                    "Content-Type": "application/json",
                },
            // redirect: "follow", // manual, *follow, error
            // referrerPolicy: "no-referrer", // no-referrer, *no-referrer-when-downgrade, origin, origin-when-cross-origin, same-origin, strict-origin, strict-origin-when-cross-origin, unsafe-url
                body: JSON.stringify(data), // body data type must match "Content-Type" header
            });

            out.code = response.status;
            out.data = await response.json();
        } catch (error) {
            console.error("There has been a problem with your fetch operation:", error);
        }

        out.ok = (out.code >= 200 && out.code <= 299)
        if (!out.ok) {
            addToLog(`ERROR ${out.code}: ${out.data?.error}`);
        }

        return out;
    }
}

const addToLog = (msg) => {
    const el = document.getElementById("log");
    const nowPrintable = formatTime(new Date());

    el.innerHTML += `${nowPrintable} ${msg}<br>`;
    el.scrollTop = el.scrollHeight;
}

const formatTime = (date) => {
    return `${date.getHours()}`.padStart(2,'0')
        + ':' + `${date.getMinutes()}`.padStart(2,'0')
        + ':' + `${date.getSeconds()}`.padStart(2,'0');
}