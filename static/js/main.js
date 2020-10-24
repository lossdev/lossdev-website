var dark = true;
var scrollLettering = 0;

window.onload = function() {
    updateLabel();
    var t = anime.timeline ({
        loop: true,
        loopComplete() {
            updateLabel();
        }
    });
    t.add({
        targets: '.fading-text-wrapper .letter',
        opacity: [0,1],
        easing: "easeOutExpo",
        duration: 1000,
        offset: '-=775',
        delay: (el, i) => 34 * (i+1)
    }).add({
        targets: '.fading-text-wrapper',
        opacity: 0,
        duration: 600,
        easing: "easeOutExpo",
        delay: 1000,
    });
}

function updateLabel() {
    var textWrapper = document.querySelector('.fading-text-wrapper .letters');
    textWrapper.innerHTML = textWrapper.textContent.replace(/([^\x00-\x80]|\w)/g, "<span class='letter'>$&</span>");
    switch(scrollLettering) {
        case 0:
            textWrapper.innerHTML = '<i class="fas fa-microchip"></i>  Software Engineer';
            scrollLettering++;
            break;
        case 1:
            textWrapper.innerHTML = '<i class="fab fa-docker"></i>  Container Escaper';
            scrollLettering++;
            break;
        case 2:
            textWrapper.innerHTML = '<i class="fas fa-code"></i>  Full Stack Web Developer';
            scrollLettering++;
            break;
        case 3:
            textWrapper.innerHTML = '<i class="fas fa-bug"></i>  Security Analyst';
            scrollLettering = 0;
            break;
    }
}

function darkMode() {
    var b = document.getElementById("body");
    var m = document.getElementById("moonIcon");
    var s = document.getElementById("lightIcon");
    if (dark) {
        b.style["background-color"] = "WhiteSmoke";
        m.style["display"] = "initial";
        s.style["display"] = "none";
        var h = document.querySelector(".header-link");
        h.style["color"] = "#1a1a1a";
        h.onmouseover = function() {
            this.style.color = "#2eabff";
        }
        h.onmouseout = function() {
            this.style.color = "#1a1a1a";
        }
        h.style.boxShadow = "0 0 1pt 1pt #1a1a1a";
        var t = document.querySelectorAll(".text");
        for (var i = 0; i < t.length; i++) {
            t[i].style["color"] = "#1a1a1a";
        }
        var l = document.querySelectorAll(".pagelink");
        for (var i = 0; i < l.length; i++) {
            l[i].style["color"] = "#1a1a1a";
            l[i].onmouseover = function() {
                this.style.color = "#2eabff";
            }
            l[i].onmouseout = function() {
                this.style.color = "#1a1a1a";
            }
        }
        dark = false;
    } else {
        b.style["background-color"] = "#333";
        m.style["display"] = "none";
        s.style["display"] = "initial";
        var h = document.querySelector(".header-link");
        h.style["color"] = "FloralWhite";
        h.onmouseover = function() {
            this.style.color = "#2eabff";
        }
        h.onmouseout = function() {
            this.style.color = "FloralWhite";
        }
        h.style.boxShadow = "0 0 1pt 1pt FloralWhite";
        var t = document.querySelectorAll(".text");
        for (var i = 0; i < t.length; i++) {
            t[i].style["color"] = "FloralWhite";
        }
        var l = document.querySelectorAll(".pagelink");
        for (var i = 0; i < l.length; i++) {
            l[i].style["color"] = "FloralWhite";
            l[i].onmouseover = function() {
                this.style.color = "#2eabff";
            }
            l[i].onmouseout = function() {
                this.style.color = "FloralWhite";
            }
        }
        dark = true;
    }
}