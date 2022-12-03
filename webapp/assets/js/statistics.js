import FilteredTable from "../components/filtered-table/filtered-table.js";

window.addEventListener('DOMContentLoaded', (e) => {

    console.log("users: ", users);

    Telegram.WebApp.expand();
    //
    // let data = [
    //     {
    //         "name": "test",
    //         "eventsNum": 5
    //     },
    //     {
    //         "name": "test",
    //         "eventsNum": 10
    //     },
    //     {
    //         "name": "test",
    //         "eventsNum": 15
    //     }
    // ];

    let table = document.getElementById("filtered-table");

    // let roles = document.getElementById("roles");

    new FilteredTable(table, users, {
        templateFunction: (user) => {
            if (user.events) {
                return `<td>${user.name}</td><td>${user.events.length}</td>`;
            }
        },
    })
})