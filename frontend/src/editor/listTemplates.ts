export type ListTemplateInsertion = {
	text: string;
	cursorLineOffset: number;
	cursorColumn: number;
};

export function listNextItemCompletion(lines: string[], lineNumber: number): ListTemplateInsertion | null {
	const currentIndex = lineNumber - 1;
	const previousIndex = currentIndex - 1;
	if (previousIndex < 0 || currentIndex >= lines.length) {
		return null;
	}

	const currentLine = lines[currentIndex] ?? "";
	if (currentLine.trim() !== "") {
		return null;
	}

	const itemStartIndex = findListItemStart(lines, previousIndex);
	if (itemStartIndex < 0 || hasNextListItem(lines, currentIndex, indentationLength(lines[itemStartIndex] ?? ""))) {
		return null;
	}

	const text = listItemTemplate(lines, itemStartIndex, previousIndex);
	if (text === "") {
		return null;
	}

	return {
		text,
		cursorLineOffset: 0,
		cursorColumn: firstEditableColumn(text),
	};
}

function findListItemStart(lines: string[], previousIndex: number): number {
	for (let index = previousIndex; index >= 0; index--) {
		const line = lines[index] ?? "";
		if (line.trim() === "") {
			return -1;
		}
		if (isListItemLine(line)) {
			return index;
		}
	}
	return -1;
}

function hasNextListItem(lines: string[], currentIndex: number, listIndent: number): boolean {
	for (let index = currentIndex + 1; index < lines.length; index++) {
		const line = lines[index] ?? "";
		if (line.trim() === "") {
			continue;
		}

		const indent = indentationLength(line);
		if (indent < listIndent) {
			return false;
		}
		if (indent === listIndent) {
			return isListItemLine(line);
		}
	}
	return false;
}

function listItemTemplate(lines: string[], itemStartIndex: number, previousIndex: number): string {
	const result: string[] = [];
	for (let index = itemStartIndex; index <= previousIndex; index++) {
		const line = lines[index] ?? "";
		if (line.trim() === "") {
			continue;
		}

		const nextLine = nextNonBlankLine(lines, index + 1, previousIndex);
		const hasNestedValue = nextLine !== null && indentationLength(nextLine) > indentationLength(line);
		result.push(emptyValueLine(line, hasNestedValue));
	}
	return result.join("\n");
}

function nextNonBlankLine(lines: string[], startIndex: number, endIndex: number): string | null {
	for (let index = startIndex; index <= endIndex; index++) {
		const line = lines[index] ?? "";
		if (line.trim() !== "") {
			return line;
		}
	}
	return null;
}

function emptyValueLine(line: string, hasNestedValue: boolean): string {
	const indent = leadingWhitespace(line);
	const trimmed = line.trim();
	if (trimmed.startsWith("- ")) {
		return `${indent}- ${emptyValue(trimmed.slice(2), hasNestedValue)}`;
	}
	return `${indent}${emptyValue(trimmed, hasNestedValue)}`;
}

function emptyValue(value: string, hasNestedValue: boolean): string {
	const colonIndex = value.indexOf(":");
	if (colonIndex < 0) {
		return "";
	}

	const key = value.slice(0, colonIndex + 1);
	return hasNestedValue ? key : `${key} `;
}

function firstEditableColumn(text: string): number {
	const firstLine = text.split("\n", 1)[0] ?? "";
	return firstLine.length + 1;
}

function isListItemLine(line: string): boolean {
	return /^\s*-\s+/.test(line);
}

function leadingWhitespace(value: string): string {
	return value.match(/^\s*/)?.[0] ?? "";
}

function indentationLength(value: string): number {
	return leadingWhitespace(value).length;
}
