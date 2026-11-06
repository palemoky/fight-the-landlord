/// <reference types="vite/client" />

declare module '*.proto?raw' {
  const source: string;
  export default source;
}
