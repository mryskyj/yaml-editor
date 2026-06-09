export type ScheduleTemplateInsertion = {
	text: string;
	cursorLineOffset: number;
	cursorColumn: number;
};

export const defaultScheduleTemplate = [
	"run1: &run1 1 #BOD",
	"run2: &run2 2 #あいうえお",
	"run3: &run3 3 #かきくけこ",
].join("\n");

export function sanitizeScheduleTemplate(value: string): string {
	const lines = value.replace(/\r\n/g, "\n").split("\n");
	while (lines.length > 0 && lines[0].trim() === "") {
		lines.shift();
	}
	while (lines.length > 0 && lines[lines.length - 1].trim() === "") {
		lines.pop();
	}

	const nonEmptyLines = lines.filter((line) => line.trim() !== "");
	const commonIndent = nonEmptyLines.reduce((indent, line) => {
		const length = leadingWhitespace(line).length;
		return indent === null ? length : Math.min(indent, length);
	}, null as number | null) ?? 0;

	return lines.map((line) => line.slice(commonIndent)).join("\n");
}

export function scheduleTemplateInsertion(
	lines: string[],
	lineNumber: number,
	template: string,
): ScheduleTemplateInsertion | null {
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
	const indent = currentLine.length > 0
		? leadingWhitespace(currentLine)
		: `${leadingWhitespace(previousLine)}    `;
	if (previousLine.trim() !== "schedules:" || hasScheduleEntries(lines, currentIndex, indent.length)) {
		return null;
	}

	const text = sanitizeScheduleTemplate(template)
		.split("\n")
		.filter((line) => line.trim() !== "")
		.map((line) => `${indent}${line}`)
		.join("\n");

	if (text === "") {
		return null;
	}

	return {
		text,
		cursorLineOffset: text.split("\n").length,
		cursorColumn: indent.length + 1,
	};
}

function hasScheduleEntries(lines: string[], currentIndex: number, entryIndentLength: number): boolean {
	for (let index = currentIndex + 1; index < lines.length; index++) {
		const line = lines[index] ?? "";
		if (line.trim() === "") {
			continue;
		}

		const indentLength = indentationLength(line);
		if (indentLength < entryIndentLength) {
			return false;
		}
		return indentLength === entryIndentLength && /^run\d+:\s+/.test(line.trim());
	}
	return false;
}

function leadingWhitespace(value: string): string {
	return value.match(/^\s*/)?.[0] ?? "";
}

function indentationLength(value: string): number {
	return leadingWhitespace(value).length;
}
