function verticalMenu() {
    var element = document.getElementById("topnav")

    if (element.className == "topnav") {
        element.className += " burger"
        return
    } 
    
    element.className = "topnav"
} 

window.addEventListener('resize', resize);
resize()
function resize(e){
    var width = (window.innerWidth > 0) ? window.innerWidth : screen.width;

    if(width > 900){
        document.getElementById("mobile-content").setAttribute("style", "display: none;")
        document.getElementById("desktop-content").setAttribute("style", "display: block;")
    }
    else{
        document.getElementById("desktop-content").setAttribute("style", "display: none;")
        document.getElementById("mobile-content").setAttribute("style", "display: block;")
    }
}
