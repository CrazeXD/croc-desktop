import { Install, CheckInstall } from "../wailsjs/go/main/Install"
import "./app.css"


function runInstall(filepath) {
    Install(filepath);
}

function checkInstall() {
    console.log("Checking install")
    installed = CheckInstall();
    if (installed) {
        console.log("Croc installed")
    }
    else {
        console.log("Croc not installed")
        // Unhide the #installpopup
        document.getElementById("installpopup").style.display = "block";
    }
}