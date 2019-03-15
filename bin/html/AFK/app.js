var linkList = document.querySelectorAll('a');
var file_name = GetFileName(window.location.href);
if (file_name.includes('.')) {
    file_name = file_name.substring(0, file_name.lastIndexOf('.'));
}
if (file_name.includes('?')) {
    file_name = file_name.substring(0, file_name.lastIndexOf('.'));
}
linkList.forEach(aElement => {
    console.log(GetFileName(aElement.href));
    if (GetFileName(aElement.href) == file_name) {
        aElement.className = "link selected";
    } else {
        aElement.className = "link non-selected";
    }
})

function GetFileName(string) {
    return string.substring(string.lastIndexOf('/') + 1);
}

function SetSelectAddr(form) {
    var select = form.getElementsByTagName("select")[0];
    form.action=select.options[select.selectedIndex].value;
    // window.alert(form.action)
}
function SetSelectAddr2(form) {
    var list = form.getElementsByTagName("select");
    var addr = list[0];
    var key = list[1];
    form.action=addr.options[addr.selectedIndex].value+key.options[key.selectedIndex].value;
    // window.alert(form.action)
}