import { useState } from "react";

export type SchemaField = {
  name: string;
  type: string;
  required: boolean;
  description?: string;
  default?: string;
  enum?: string[];
  children?: SchemaField[];
  item?: SchemaField;
  mapValue?: SchemaField;
};

type SchemaPaneProps = {
  root: SchemaField;
  content: string;
  cursor: {
    line: number;
    column: number;
  };
};

type TabID = "current" | "search" | "tree";

export function SchemaPane({ root, content, cursor }: SchemaPaneProps) {
  const [activeTab, setActiveTab] = useState<TabID>("current");
  const [query, setQuery] = useState("");
  const context = schemaContext(root, content, cursor.line);
  const searchResults = query.trim() ? searchSchema(root, query.trim()) : [];

  return (
    <aside className="schema-pane">
      <div className="schema-pane-header">
        <h2>Schema</h2>
        <div className="schema-tabs" role="tablist" aria-label="schema views">
          <button
            className={activeTab === "current" ? "schema-tab active" : "schema-tab"}
            type="button"
            onClick={() => setActiveTab("current")}
          >
            Current
          </button>
          <button
            className={activeTab === "search" ? "schema-tab active" : "schema-tab"}
            type="button"
            onClick={() => setActiveTab("search")}
          >
            Search
          </button>
          <button
            className={activeTab === "tree" ? "schema-tab active" : "schema-tab"}
            type="button"
            onClick={() => setActiveTab("tree")}
          >
            Tree
          </button>
        </div>
      </div>
      {activeTab === "current" ? <CurrentSchemaView context={context} /> : null}
      {activeTab === "search" ? (
        <SearchSchemaView query={query} results={searchResults} onQueryChange={setQuery} />
      ) : null}
      {activeTab === "tree" ? <SchemaTree root={root} /> : null}
    </aside>
  );
}

type SchemaContext = {
  rootName: string;
  path: string[];
  current: SchemaField;
  siblings: CandidateField[];
};

type CandidateField = {
  field: SchemaField;
  used: boolean;
};

type SearchResult = {
  field: SchemaField;
  path: string[];
};

function CurrentSchemaView({ context }: { context: SchemaContext }) {
  const breadcrumb = [context.rootName, ...context.path].filter(Boolean).join(" > ");
  return (
    <div className="schema-current">
      <div className="schema-breadcrumb" title={breadcrumb || context.current.name}>
        {breadcrumb || context.current.name}
      </div>
      <FieldSummary field={context.current} />
      <section className="schema-section">
        <h3>Available keys</h3>
        <div className="schema-candidate-list">
          {context.siblings.length === 0 ? (
            <p className="schema-empty">No child keys for this position.</p>
          ) : (
            context.siblings.map(({ field, used }) => (
              <div className={used ? "schema-candidate used" : "schema-candidate"} key={field.name}>
                <div className="schema-candidate-main">
                  <span className="schema-key">{field.name}</span>
                  <span className="schema-required">{field.required ? "required" : "optional"}</span>
                  {used ? <span className="schema-used">used</span> : null}
                </div>
                {field.description ? (
                  <div className="schema-candidate-description">{field.description}</div>
                ) : null}
              </div>
            ))
          )}
        </div>
      </section>
    </div>
  );
}

function SearchSchemaView({
  query,
  results,
  onQueryChange,
}: {
  query: string;
  results: SearchResult[];
  onQueryChange: (query: string) => void;
}) {
  return (
    <div className="schema-search">
      <input
        className="schema-search-input"
        type="search"
        value={query}
        placeholder="Search keys or descriptions"
        onChange={(event) => onQueryChange(event.target.value)}
      />
      <div className="schema-search-results">
        {query.trim() && results.length === 0 ? <p className="schema-empty">No matches.</p> : null}
        {results.map((result) => (
          <div className="schema-search-result" key={result.path.join(".")}>
            <div className="schema-search-path">{result.path.join(" > ")}</div>
            <FieldSummary field={result.field} compact />
          </div>
        ))}
      </div>
    </div>
  );
}

function SchemaTree({ root }: { root: SchemaField }) {
  return (
    <div className="schema-tree">
      {containerChildren(root).map((field) => (
        <SchemaNode key={field.name} field={field} depth={0} />
      ))}
    </div>
  );
}

function FieldSummary({ field, compact = false }: { field: SchemaField; compact?: boolean }) {
  return (
    <section className={compact ? "schema-summary compact" : "schema-summary"}>
      <div className="schema-primary">
        <span className="schema-key">{field.name || "root"}</span>
        {field.description ? <span className="schema-description">{field.description}</span> : null}
      </div>
      <div className="schema-details">
        <span>{field.type}</span>
        <span>{field.required ? "required" : "optional"}</span>
        {field.default ? <span>default {field.default}</span> : null}
        {field.enum && field.enum.length > 0 ? <span>{field.enum.join(", ")}</span> : null}
      </div>
    </section>
  );
}

function SchemaNode({ field, depth }: { field: SchemaField; depth: number }) {
  return (
    <>
      <div className="schema-node" style={{ paddingLeft: depth * 18 }}>
        <FieldSummary field={field} compact />
      </div>
      {containerChildren(field).map((child) => (
        <SchemaNode key={`${field.name}.${child.name}`} field={child} depth={depth + 1} />
      ))}
    </>
  );
}

function schemaContext(root: SchemaField, content: string, cursorLine: number): SchemaContext {
  const path = inferPath(content, cursorLine);
  const current = fieldAtPath(root, path) ?? root;
  const parentPath = fieldAtPath(root, path) ? path.slice(0, -1) : path;
  const parent = fieldAtPath(root, parentPath) ?? root;
  const used = usedKeysAtPath(content, parentPath);

  return {
    rootName: root.name,
    path: fieldAtPath(root, path) ? path : parentPath,
    current,
    siblings: containerChildren(parent).map((field) => ({
      field,
      used: used.has(field.name),
    })),
  };
}

function inferPath(content: string, cursorLine: number): string[] {
  const stack: Array<{ indent: number; key: string }> = [];
  const lines = content.split(/\r?\n/).slice(0, Math.max(cursorLine, 1));

  for (const line of lines) {
    const entry = parseKeyLine(line);
    if (!entry) {
      continue;
    }
    while (stack.length > 0 && stack[stack.length - 1].indent >= entry.indent) {
      stack.pop();
    }
    stack.push(entry);
  }

  return stack.map((entry) => entry.key);
}

function usedKeysAtPath(content: string, targetPath: string[]): Set<string> {
  const used = new Set<string>();
  const stack: Array<{ indent: number; key: string }> = [];

  for (const line of content.split(/\r?\n/)) {
    const entry = parseKeyLine(line);
    if (!entry) {
      continue;
    }
    while (stack.length > 0 && stack[stack.length - 1].indent >= entry.indent) {
      stack.pop();
    }
    if (samePath(stack.map((item) => item.key), targetPath)) {
      used.add(entry.key);
    }
    stack.push(entry);
  }

  return used;
}

function parseKeyLine(line: string): { indent: number; key: string } | null {
  const match = line.match(/^(\s*)(?:-\s*)?([A-Za-z0-9_-]+)\s*:/);
  if (!match) {
    return null;
  }
  return {
    indent: match[1].length,
    key: match[2],
  };
}

function fieldAtPath(root: SchemaField, path: string[]): SchemaField | null {
  let current: SchemaField = root;
  for (const key of path) {
    const child = containerChildren(current).find((field) => field.name === key);
    if (!child) {
      return null;
    }
    current = child;
  }
  return current;
}

function containerChildren(field: SchemaField): SchemaField[] {
  if (field.children && field.children.length > 0) {
    return field.children;
  }
  if (field.item) {
    return containerChildren(field.item);
  }
  if (field.mapValue) {
    return containerChildren(field.mapValue);
  }
  return [];
}

function searchSchema(root: SchemaField, query: string): SearchResult[] {
  const lowerQuery = query.toLowerCase();
  const results: SearchResult[] = [];

  walkSchema(root, [], (field, path) => {
    const haystack = [field.name, field.type, field.description, field.default, ...(field.enum ?? [])]
      .filter(Boolean)
      .join(" ")
      .toLowerCase();
    if (haystack.includes(lowerQuery)) {
      results.push({ field, path });
    }
  });

  return results.slice(0, 80);
}

function walkSchema(
  field: SchemaField,
  path: string[],
  visit: (field: SchemaField, path: string[]) => void,
) {
  const nextPath = field.name ? [...path, field.name] : path;
  visit(field, nextPath);
  for (const child of containerChildren(field)) {
    walkSchema(child, nextPath, visit);
  }
}

function samePath(left: string[], right: string[]): boolean {
  return left.length === right.length && left.every((item, index) => item === right[index]);
}
