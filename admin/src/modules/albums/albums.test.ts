import { resourceDefinition } from './resource';

// This contract check uses the existing Node type/runtime surface; the generator
// does not add a frontend test-runner dependency.
if (resourceDefinition.name !== "albums") throw new Error('generated resource name mismatch');
if (resourceDefinition.route !== "/albums") throw new Error('generated resource route mismatch');
if (JSON.stringify(resourceDefinition.actions) !== JSON.stringify([{ name: "view", label: "View", kind: "", permission: "admin.albums.view", batch: false, payload: "" }, { name: "create", label: "Create", kind: "", permission: "admin.albums.create", batch: false, payload: "" }, { name: "update", label: "Update", kind: "", permission: "admin.albums.update", batch: false, payload: "" }, { name: "delete", label: "Delete", kind: "", permission: "admin.albums.delete", batch: false, payload: "" }])) throw new Error('generated resource actions mismatch');
