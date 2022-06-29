function createInfoMessage(message){
    var alert = document.getElementById("info")
    var clone = alert.content.cloneNode(true)

    var messages = document.getElementById("messages")
    messages.appendChild(clone)

    text = document.getElementsByClassName("infotext")
    text[text.length-1].innerText = message
    
    validated = false
}

validated = true

function validate(){
    validated = true
    var messages = document.getElementsByClassName("info")

    for(let i=messages.length-1; i>=0; i--){
        messages[i].parentNode.removeChild(messages[i])
    }
    
    var inputs = {
        "firstname" : document.getElementById("firstname"),
        "lastname" : document.getElementById("lastname"),
        "username" : document.getElementById("username"),
        "password" : document.getElementById("password"),
        "email" : document.getElementById("email"),
    }

    for (const [ key, value ] of Object.entries(inputs)) {
        if (value.value == 0){
            createInfoMessage(key + " field is empty.")
        }
    }

    if (document.getElementById("dob-day").value == "Day" ||
        document.getElementById("dob-month").value == "Month" ||
        document.getElementById("dob-year").value == "Year"
        ){
        createInfoMessage("Please enter your date of birth correctly")
    }

    var email = inputs["email"].value
    var contains_at = false
    var contains_dot = false
    for (let i = 0; i < email.length; i++) {
        if (email[i] == "@"){
            contains_at = true
        }
        if (email[i] == "."){
            contains_dot = true
        }
    }
    
    if((!contains_at || !contains_dot) && inputs["email"].value != 0){
        createInfoMessage("Please enter email correctly")
    }

    if(validated){
        var form = document.getElementById("signupform")
        form.submit()
    }
    
}
