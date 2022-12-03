class FilteredTable {
    constructor(table, data, options) {
        this.data = data;
        this.options = options;

        this.elements = {
            table: table,
            body: document.createElement("tbody")
        };

        this.elements.body.id = "filtered-table-body"
        this.elements.body.classList.add(
            "filtered-table__body"
        );
        this.elements.table.appendChild(this.elements.body);

        this.addListeners()

        console.log(this.data)
        this.populateResults(this.data)
    }

    addListeners() {
        this.options.customListeners(this).forEach(listener => {
            listener.element.addEventListener(listener.type, listener.func)
        })

        this.options.filters(this).forEach(filter => {
            filter.element.addEventListener(filter.type, filter.func)
        })
    }

    populateResults(results) {
        // Clear all existing results
        while (this.elements.body.firstChild) {
            this.elements.body.removeChild(
                this.elements.body.firstChild
            );
        }

        results = results.filter(r => r.events && r.events.length > 0).sort((r1, r2) => r1.events.length < r2.events.length);

        // console.log(results)
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