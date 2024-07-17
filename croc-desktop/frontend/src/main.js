import { Install, CheckInstall } from "../wailsjs/go/main/Install"
import { Quit } from "../wailsjs/go/main/App"
import { OpenFile } from "../wailsjs/go/main/Croc"
import "./app.css"

let installed = false;
// On page load, check if Croc is installed
document.addEventListener("DOMContentLoaded", async function () {
    let installed = await checkInstall();

    if (!installed) {
        document.getElementById("installpopup-no").addEventListener("click", Quit);
        document.getElementById("installpopup-yes").addEventListener("click", () => runInstall("croc"));
    }
    const installPopup = document.getElementById('installpopup');


    // Check if croc is installed (you'll need to implement this check)
    let isCrocInstalled = installed; // Replace with actual check
    const content = document.getElementById('content');
    if (!isCrocInstalled) {
        installPopup.style.display = 'block';
    }
    else {
        content.classList.remove('slide-transition');
        content.classList.add('slide-up');
    }
});

async function runInstall(filepath) {
    await Install(filepath);
    console.log("Install complete")
    document.getElementById("installpopup").style.display = "none";
    installed = true;
    const content = document.getElementById('content');
    content.classList.add('slide-up');
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

document.getElementById("send").addEventListener("click", async function () {
    await OpenFile();
}
);