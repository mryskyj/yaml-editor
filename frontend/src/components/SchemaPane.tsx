export type SchemaField = {
  name: string;
  type: string;
  required: boolean;
  description?: string;
  default?: string;
  enum?: string[];
  children?: SchemaField[];
};

type SchemaPaneProps = {
  root: SchemaField;
};

export function SchemaPane({ root }: SchemaPaneProps) {
  return (
    <aside className="schema-pane">
      <h2>Schema</h2>
      <div className="schema-tree">
        {root.children?.map((field) => (
          <SchemaNode key={field.name} field={field} depth={0} />
        ))}
      </div>
    </aside>
  );
}

function SchemaNode({ field, depth }: { field: SchemaField; depth: number }) {
  return (
    <>
      <div className="schema-node" style={{ paddingLeft: depth * 18 }}>
        <div className="schema-primary">
          <span className="schema-key">{field.name}</span>
          {field.description ? <span className="schema-description">{field.description}</span> : null}
        </div>
        <div className="schema-details">
          <span>{field.type}</span>
          <span>{field.required ? "required" : "optional"}</span>
          {field.default ? <span>default {field.default}</span> : null}
          {field.enum && field.enum.length > 0 ? <span>{field.enum.join(", ")}</span> : null}
        </div>
      </div>
      {field.children?.map((child) => (
        <SchemaNode key={`${field.name}.${child.name}`} field={child} depth={depth + 1} />
      ))}
    </>
  );
}
