export type ListTemplateInsertion = {
	label: string;
	text: string;
	cursorLineOffset: number;
	cursorColumn: number;
	insertLineNumber?: number;
};

export function listNextItemCompletion(lines: string[], lineNumber: number): ListTemplateInsertion | null {
	return listNextItemCompletions(lines, lineNumber)[0] ?? null;
}

export function listNextItemCompletions(lines: string[], lineNumber: number): ListTemplateInsertion[] {
	const currentIndex = lineNumber - 1;
	const previousIndex = currentIndex - 1;
	if (previousIndex < 0 || currentIndex >= lines.length) {
		return [];
	}

	const currentLine = lines[currentIndex] ?? "";
	if (currentLine.trim() !== "") {
		return [];
	}

	const completions: ListTemplateInsertion[] = [];
	for (const itemStartIndex of findListItemStarts(lines, previousIndex)) {
		const text = listItemTemplate(lines, itemStartIndex, previousIndex);
		if (text === "") {
			continue;
		}

		completions.push({
			label: `Add ${listParentName(lines, itemStartIndex)} item`,
			text,
			cursorLineOffset: 0,
			cursorColumn: firstEditableColumn(text),
			insertLineNumber: listInsertionLineNumber(lines, itemStartIndex),
		});
	}
	return completions;
}

function findListItemStarts(lines: string[], previousIndex: number): number[] {
	const starts: number[] = [];
	let maxIndent = Number.POSITIVE_INFINITY;
	for (let index = previousIndex; index >= 0; index--) {
		const line = lines[index] ?? "";
		if (line.trim() === "") {
			break;
		}

		const indent = indentationLength(line);
		if (indent >= maxIndent) {
			continue;
		}
		if (isListItemLine(line)) {
			starts.push(index);
			maxIndent = indent;
		}
	}
	return starts;
}

function findListItemStart(lines: string[], previousIndex: number): number {
	return findListItemStarts(lines, previousIndex)[0] ?? -1;
}

function listInsertionLineNumber(lines: string[], itemStartIndex: number): number {
	const listIndent = indentationLength(lines[itemStartIndex] ?? "");
	for (let index = itemStartIndex + 1; index < lines.length; index++) {
		const line = lines[index] ?? "";
		if (line.trim() === "") {
			continue;
		}

		const indent = indentationLength(line);
		if (indent <= listIndent) {
			return index + 1;
		}
	}
	return lines.length + 1;
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

function listParentName(lines: string[], itemStartIndex: number): string {
	const listIndent = indentationLength(lines[itemStartIndex] ?? "");
	for (let index = itemStartIndex - 1; index >= 0; index--) {
		const line = lines[index] ?? "";
		if (line.trim() === "") {
			continue;
		}

		const indent = indentationLength(line);
		if (indent >= listIndent) {
			continue;
		}

		const key = yamlKey(line.trim());
		return key || "list";
	}
	return "list";
}

function yamlKey(trimmedLine: string): string {
	const value = trimmedLine.startsWith("- ") ? trimmedLine.slice(2).trimStart() : trimmedLine;
	const colonIndex = value.indexOf(":");
	if (colonIndex <= 0) {
		return "";
	}
	return value.slice(0, colonIndex).trim();
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
