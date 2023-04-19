import * as d3 from "d3";

class Shell {
    constructor(target, handleResponse) {
        this.target = d3.select(`#${target}`);
        this.handleResponse = handleResponse;
        this.items = this.target.append("div").attr("class", "items");
        this.closed = true;
        this.history = -1;
        this.changes = [];
        this.lines = [[{Text: "b6 from "}, {Text: "Diagonal"}]];
        const form = this.items.append("form");
        form.append("span").attr("class", "prompt").text("b6");
        this.input = form.append("input").attr("type", "text").attr("autocomplete", "off");
        this.input.on("keydown", (e) => {
            if (e.key == "ArrowUp") {
                e.preventDefault();
                this.copyNextLineFromHistory(true);
            } else if (e.key == "ArrowDown") {
                e.preventDefault();
                this.copyNextLineFromHistory(false);
            }
        });
        form.on("submit", (e) => {
            e.preventDefault();
            this.evaluate();
        });
        this.render();
    }

    toggle() {
        this.closed = !this.closed;
        this.target.attr("class", this.closed ? "closed" : "open");
        if (!this.closed) {
            this.input.node().focus();
        }
    }

    evaluateExpression(e) {
        this.input.node().value = e;
        this.evaluate();
    }

    evaluate() {
        const request = {
            method: "POST",
            body: JSON.stringify({
                e: this.input.node().value,
                cs: this.changes
            }),
            headers: {
                "Content-type": "application/json; charset=UTF-8"
            }
        };
        d3.json("/shell", request).then(response => {
            this.lines = this.lines.concat(response.Lines);
            this.history = -1;
            this.handleResponse(response);
            this.render();
            this.input.node().focus();
        });
        this.input.node().value = "";

    }

    render() {
        const lines = this.items.selectAll(".line").data(this.lines).join("div");
        lines.attr("class", "line");
        const spans = lines.selectAll("span").data(l => l).join("span");
        spans.attr("class", s => s.Class != "" ? s.Class : null).text(s => s.Text);
        const literals = lines.selectAll(".literal").on("click", (e, s) => {
            const value = this.input.node().value;
            if (value && !value.endsWith(" ")) {
                this.input.node().value = value + " ";
            }
            this.input.node().value = this.input.node().value + s.Text + " ";
            this.input.node().focus();
        });
        this.items.select("form").raise();
    }

    copyNextLineFromHistory(backwards) {
        const delta = backwards ? 1 : -1;
        const origin = this.history;
        while (this.history + delta >= -1 && this.history + delta <= this.lines.length) {
            this.history += delta;
            if (this.history >= 0 && this.history < this.lines.length) {
                const line = this.lines[this.lines.length - this.history - 1];
                if (line.length > 0 && line[0].Class == "prompt") {
                    break;
                }
            }
        }
        if (this.history >= this.lines.length) {
            this.history = origin;
        }
        if (this.history < 0) {
            this.input.node().value = "";
        } else {
            let value = "";
            const line = this.lines[this.lines.length - this.history - 1];
            for (const i in line) {
                const s = line[i];
                if (s.Class != "prompt" && s.Text) {
                    value += s.Text;
                }
            }
            this.input.node().value = value;
        }
    }
};

export default Shell;