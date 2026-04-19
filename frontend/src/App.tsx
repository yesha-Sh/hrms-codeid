import { AuthProvider } from "./auth/AuthContext";
import { AppRouter } from "./router/AppRouter";

export default function App() {
  return (
    <AuthProvider>
      <AppRouter />
    </AuthProvider>
  );
}
