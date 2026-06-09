export type DateTemplateInsertion = {
	text: string;
	cursorLineOffset: number;
	cursorColumn: number;
};

const dateValuePattern = /^date:\s*"?(\d{4}-\d{2}-\d{2})"?\s*$/;
const dayKeyPattern = /^day(\d+):\s*$/;
const holidayPattern = /^holiday:\s*(true|false)\s*$/;

export function dateTemplateInsertion(lines: string[], lineNumber: number): DateTemplateInsertion | null {
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
	if (previousLine.trim() === "dates:") {
		const indent = currentLine.length > 0
			? leadingWhitespace(currentLine)
			: `${leadingWhitespace(previousLine)}    `;
		return dateBlockInsertion(indent, 1, "");
	}

	return null;
}

export function dateNextBlockCompletion(lines: string[], lineNumber: number): DateTemplateInsertion | null {
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
	if (!holidayPattern.test(previousLine.trim())) {
		return null;
	}

	const dayContext = findDateDayContext(lines, previousIndex);
	if (!dayContext || hasNextDay(lines, currentIndex, dayContext.dayIndent)) {
		return null;
	}

	return dateBlockInsertion(
		dayContext.dayIndent,
		dayContext.dayNumber + 1,
		nextDate(dayContext.date),
	);
}

function dateBlockInsertion(indent: string, dayNumber: number, date: string): DateTemplateInsertion {
	const childIndent = `${indent}    `;
	return {
		text: `${indent}day${dayNumber}:\n${childIndent}date: "${date}"\n${childIndent}holiday: false`,
		cursorLineOffset: 1,
		cursorColumn: childIndent.length + `date: "`.length + date.length + 1,
	};
}

function findDateDayContext(lines: string[], holidayIndex: number): {
	dayIndent: string;
	dayNumber: number;
	date: string;
} | null {
	const holidayIndentLength = indentationLength(lines[holidayIndex] ?? "");
	let date = "";

	for (let index = holidayIndex - 1; index >= 0; index--) {
		const line = lines[index] ?? "";
		const trimmed = line.trim();
		const indentLength = indentationLength(line);

		if (indentLength === holidayIndentLength) {
			const match = trimmed.match(dateValuePattern);
			if (match) {
				date = match[1];
			}
			continue;
		}

		if (indentLength >= holidayIndentLength) {
			continue;
		}

		const dayMatch = trimmed.match(dayKeyPattern);
		if (!dayMatch) {
			continue;
		}

		if (!isUnderDates(lines, index, indentLength)) {
			return null;
		}

		return {
			dayIndent: leadingWhitespace(line),
			dayNumber: Number(dayMatch[1]),
			date,
		};
	}

	return null;
}

function isUnderDates(lines: string[], dayIndex: number, dayIndentLength: number): boolean {
	for (let index = dayIndex - 1; index >= 0; index--) {
		const line = lines[index] ?? "";
		if (line.trim() === "") {
			continue;
		}

		const indentLength = indentationLength(line);
		if (indentLength >= dayIndentLength) {
			continue;
		}

		return line.trim() === "dates:";
	}
	return false;
}

function hasNextDay(lines: string[], currentIndex: number, dayIndent: string): boolean {
	for (let index = currentIndex + 1; index < lines.length; index++) {
		const line = lines[index] ?? "";
		if (line.trim() === "") {
			continue;
		}

		if (indentationLength(line) < dayIndent.length) {
			return false;
		}
		if (leadingWhitespace(line) === dayIndent) {
			return dayKeyPattern.test(line.trim());
		}
	}
	return false;
}

function nextDate(value: string): string {
	if (!dateValuePattern.test(`date: "${value}"`)) {
		return "";
	}

	const date = new Date(`${value}T00:00:00Z`);
	if (Number.isNaN(date.getTime())) {
		return "";
	}

	date.setUTCDate(date.getUTCDate() + 1);
	return date.toISOString().slice(0, 10);
}

function leadingWhitespace(value: string): string {
	return value.match(/^\s*/)?.[0] ?? "";
}

function indentationLength(value: string): number {
	return leadingWhitespace(value).length;
}
