import { Install, CheckInstall } from "../wailsjs/go/main/Install"
import { Quit } from "../wailsjs/go/main/App"
import "./app.css"

let installed = false;
// On page load, check if Croc is installed
document.addEventListener("DOMContentLoaded", async function() {
    let installed = await checkInstall();
    
    if (!installed) {
        document.getElementById("installpopup-no").addEventListener("click", Quit);
        document.getElementById("installpopup-yes").addEventListener("click", () => runInstall("croc"));
    }
});

async function runInstall(filepath) {
    await Install(filepath);
    console.log("Install complete")
    document.getElementById("installpopup").style.display = "none";
    installed = true;
}

async function checkInstall() {
    console.log("Checking install")
    installed = await CheckInstall();
    if (installed) {
        console.log("Croc installed")
        return true;
    }
    else {
        console.log("Croc not installed")
        // Unhide the #installpopup
        document.getElementById("installpopup").style.display = "block";
        return false;
    }
}