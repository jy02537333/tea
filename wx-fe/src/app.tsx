import type { PropsWithChildren } from 'react';
import './app.scss';

function App({ children }: PropsWithChildren<Record<string, unknown>>) {
  return children;
}

export default App;
