import FilteredTable from "../components/filtered-table/filtered-table.js";

window.addEventListener('DOMContentLoaded', async (e) => {

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

    let fromDate = document.getElementById("from-date")

    let users = await getUsers(bandId, moment(fromDate.valueAsDate).format('DD.MM.YYYY'))

    let table = document.getElementById("filtered-table");

    let filters = {}

    new FilteredTable(table, users, {
        templateFunction: (user) => {
            if (user.events) {
                // todo
                // let eventsStr = Array.from(new Set(user.events.map(event => event.name))).join(", ");
                // <!--<td>${eventsStr}</td>-->
                return `<td>${user.name}</td><td>${user.events.length}</td>`;
            }
        },
        customListeners: (filteredTable) => {
            return [
                {
                    type: "change",
                    element: document.getElementById("from-date"),
                    func: async (event) => {
                        if (moment(event.target.valueAsDate).isBefore(moment(initFromDate))) {

                            filteredTable.data = await getUsers(bandId, moment(event.target.valueAsDate).format('DD.MM.YYYY'));

                            console.log(filteredTable.data)
                            filteredTable.populateResults(filteredTable.data);
                        }
                    }
                },
            ]
        },
        filters: (filteredTable) => {
            return [
                {
                    type: "change",
                    element: document.getElementById("roles"),
                    func: (event) => {
                        filters["roles"] = Array.from(event.target.selectedOptions).map(option => option.value);

                        let filteredData = filterData(filteredTable.data, filters);
                        filteredTable.populateResults(filteredData);
                    }
                },
                {
                    type: "change",
                    element: document.getElementById("weekdays"),
                    func: (event) => {
                        filters["weekdays"] = Array.from(event.target.selectedOptions).map(option => option.value);

                        let filteredData = filterData(filteredTable.data, filters);
                        filteredTable.populateResults(filteredData);
                    }
                },
                {
                    type: "change",
                    element: document.getElementById("from-date"),
                    func: (event) => {
                        filters["fromDate"] = event.target.valueAsDate;

                        let filteredData = filterData(filteredTable.data, filters);
                        filteredTable.populateResults(filteredData);
                    }
                }
            ]
        }
    })
})

function filterData(data, filters) {

    let dataCopy = JSON.parse(JSON.stringify(data));
    return dataCopy
        .filter(user => user.events)
        .map(user => {
            user.events = user.events.filter(event => {

                let bools = [];

                for (const [key, chosenValues] of Object.entries(filters)) {
                    if (chosenValues.length === 0 || (chosenValues.length === 1 && chosenValues[0] === "")) {
                        bools.push(true);
                    } else {
                        switch (key) {
                            case "roles": {
                                let b = event.roles.some(role => {
                                    return chosenValues.includes(role.id);
                                });
                                bools.push(b);
                                break;
                            }
                            case "weekdays": {
                                let b = chosenValues.includes(event.weekday.toString());
                                bools.push(b);
                                break;
                            }
                            case "fromDate": {
                                const eventDate = moment(event.date);
                                const fromDate = moment(chosenValues);

                                let b = fromDate.isBefore(eventDate, "day");
                                bools.push(b);
                                break;
                            }
                        }
                    }
                }

                // console.log(bools);
                return !bools.includes(false);
            });
            return user;
        })
}

async function getUsers(bandId, fromDate) {
    let resp = await fetch(`/api/users-with-events?bandId=${bandId}&from=${fromDate}`, {
        method: "get",
        headers: {'Content-Type': 'application/json'},
    })
    let data = await resp.json();
    return data.users;
}