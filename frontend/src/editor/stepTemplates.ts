export type StepTemplateInsertion = {
	text: string;
	cursorLineOffset: number;
	cursorColumn: number;
};

export function stepTemplateInsertion(lines: string[], lineNumber: number): StepTemplateInsertion | null {
	const currentIndex = lineNumber - 1;
	const previousIndex = currentIndex - 1;
	if (previousIndex < 0 || currentIndex >= lines.length) {
		return null;
	}

	const currentLine = lines[currentIndex] ?? "";
	if (currentLine.trim() !== "") {
		return null;
	}

	const previousLine = lines[previousIndex] ?? "";
	const entryIndent = currentLine.length > 0
		? leadingWhitespace(currentLine)
		: `${leadingWhitespace(previousLine)}  `;
	if (previousLine.trim() !== "steps:" || hasStepEntries(lines, currentIndex, entryIndent.length)) {
		return null;
	}

	const childIndent = `${entryIndent}  `;
	const actionChildIndent = `${childIndent}  `;
	const argsChildIndent = `${actionChildIndent}  `;
	const text = [
		`${entryIndent}- id: ""`,
		`${childIndent}name: ""`,
		`${childIndent}day_ref: `,
		`${childIndent}schedule_ref: `,
		`${childIndent}action:`,
		`${actionChildIndent}tool: ""`,
		`${actionChildIndent}args:`,
		`${argsChildIndent}`,
	].join("\n");

	return {
		text,
		cursorLineOffset: 1,
		cursorColumn: entryIndent.length + `- id: "`.length + 1,
	};
}

function hasStepEntries(lines: string[], currentIndex: number, entryIndentLength: number): boolean {
	for (let index = currentIndex + 1; index < lines.length; index++) {
		const line = lines[index] ?? "";
		if (line.trim() === "") {
			continue;
		}

		const indentLength = indentationLength(line);
		if (indentLength < entryIndentLength) {
			return false;
		}
		return indentLength === entryIndentLength && line.trim().startsWith("-");
	}
	return false;
}

function leadingWhitespace(value: string): string {
	return value.match(/^\s*/)?.[0] ?? "";
}

function indentationLength(value: string): number {
	return leadingWhitespace(value).length;
}
