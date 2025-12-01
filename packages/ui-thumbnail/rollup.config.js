import path from 'path';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import typescript from '@rollup/plugin-typescript';
import peerDepsExternal from 'rollup-plugin-peer-deps-external';
import { terser } from 'rollup-plugin-terser';

const pkg = require('./package.json');

export default [
  // React bundle (ESM + CJS)
  {
    input: 'react/index.tsx',
    output: [
      { file: 'dist/react/index.cjs.js', format: 'cjs', sourcemap: true },
      { file: 'dist/react/index.esm.js', format: 'es', sourcemap: true }
    ],
    plugins: [
      peerDepsExternal(),
      resolve({ extensions: ['.js', '.jsx', '.ts', '.tsx'] }),
      commonjs(),
      typescript({ tsconfig: 'react/tsconfig.json' }),
      terser()
    ]
  },
  // Taro bundle (ESM)
  {
    input: 'taro/index.tsx',
    output: [
      { file: 'dist/taro/index.esm.js', format: 'es', sourcemap: true }
    ],
    plugins: [
      peerDepsExternal(),
      resolve({ extensions: ['.js', '.jsx', '.ts', '.tsx'] }),
      commonjs(),
      typescript({ tsconfig: 'taro/tsconfig.json' }),
      terser()
    ]
  }
];
