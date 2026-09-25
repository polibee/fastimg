export * from 'vue'

export const isVue2 = false
export const isVue3 = true
export const Vue2 = undefined
export const install = () => undefined
export const set = (target, key, value) => {
  target[key] = value
}
export const del = (target, key) => {
  delete target[key]
}
