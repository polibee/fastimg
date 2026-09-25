import { resourceDefinition } from './resource';

// This contract check uses the existing Node type/runtime surface; the generator
// does not add a frontend test-runner dependency.
if (resourceDefinition.name !== "folders") throw new Error('generated resource name mismatch');
if (resourceDefinition.route !== "/folders") throw new Error('generated resource route mismatch');
if (JSON.stringify(resourceDefinition.actions) !== JSON.stringify([{ name: "view", label: "View", kind: "", permission: "admin.folders.view", batch: false, payload: "" }, { name: "create", label: "Create", kind: "", permission: "admin.folders.create", batch: false, payload: "" }, { name: "update", label: "Update", kind: "", permission: "admin.folders.update", batch: false, payload: "" }, { name: "delete", label: "Delete", kind: "", permission: "admin.folders.delete", batch: false, payload: "" }])) throw new Error('generated resource actions mismatch');
