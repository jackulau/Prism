import { RouterProvider } from 'react-router-dom';
import { router } from './routes';
import { useEffect } from 'react';
import { initAuth } from './store/authStore';

function App() {
  // Initialize auth on mount
  useEffect(() => {
    initAuth();
  }, []);

  return <RouterProvider router={router} />;
}

export default App;
