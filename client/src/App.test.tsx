import React from 'react';
import { render } from '@testing-library/react';
import App from './App';

test('renders welcome message', () => {
  const { getByText } = render(<App />);
  const linkElement = getByText(/welcome to sjles pta voting/i);
  expect(linkElement).toBeInTheDocument();
});
