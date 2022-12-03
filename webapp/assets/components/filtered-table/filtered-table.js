class FilteredTable {
    constructor(filteredTable, data, options) {
        this.data = data;
        this.options = options;

        this.elements = {
            table: filteredTable,
            body: document.createElement("tbody")
        };

        this.elements.body.id = "filtered-table-body"
        this.elements.body.classList.add(
            "filtered-table__body"
        );
        this.elements.table.appendChild(this.elements.body);

        this.addListeners()

        this.populateResults(this.data)
    }

    addListeners() {
        let roles = document.getElementById("roles");
        roles.onchange = (e) => {
            let dataCopy = JSON.parse(JSON.stringify(this.data));

            let selectedRoleIDs = Array.from(roles.selectedOptions).map(option => option.value);

            let filteredData = dataCopy.map(user => {
                if (!user.events) {
                    return user;
                }
                user.events = user.events.filter(event => {
                    return event.roles.some(role => {
                        return selectedRoleIDs.includes(role.id);
                    });
                });
                return user;
            })

            this.populateResults(filteredData)
        }
    }

    populateResults(results) {
        // Clear all existing results
        while (this.elements.body.firstChild) {
            this.elements.body.removeChild(
                this.elements.body.firstChild
            );
        }

        results = results.filter(r => r.events && r.events.length > 0).sort((r1, r2) => r1.events.length < r2.events.length);

        // Update list of results under the search bar
        for (const result of results) {
            this.elements.body.appendChild(
                this.createRow(result)
            );
        }
    }

    createRow(result) {
        const anchorElement = document.createElement("tr");

        anchorElement.classList.add("filtered-table__row");
        anchorElement.insertAdjacentHTML(
            "afterbegin",
            this.options.templateFunction(result)
        );

        // anchorElement.addEventListener(...this.options.resultEventListener(result, this))

        return anchorElement;
    }
}

export default FilteredTable;