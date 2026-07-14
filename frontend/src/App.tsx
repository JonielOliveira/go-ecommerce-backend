import { Navigate, Route, Routes } from "react-router-dom";
import { AppLayout } from "./components/AppLayout";
import { AdminRoute, ProtectedRoute } from "./components/ProtectedRoute";
import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";
import { ProductsPage } from "./pages/ProductsPage";
import { UsersPage } from "./pages/UsersPage";
import { OrdersPage } from "./pages/OrdersPage";
import { CartPage } from "./pages/CartPage";

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/registro" element={<RegisterPage />} />

      <Route
        path="/produtos"
        element={
          <ProtectedRoute>
            <AppLayout>
              <ProductsPage />
            </AppLayout>
          </ProtectedRoute>
        }
      />

      <Route
        path="/pedidos"
        element={
          <ProtectedRoute>
            <AppLayout>
              <OrdersPage />
            </AppLayout>
          </ProtectedRoute>
        }
      />

      <Route
        path="/carrinho"
        element={
          <ProtectedRoute>
            <AppLayout>
              <CartPage />
            </AppLayout>
          </ProtectedRoute>
        }
      />

      <Route
        path="/usuarios"
        element={
          <AdminRoute>
            <AppLayout>
              <UsersPage />
            </AppLayout>
          </AdminRoute>
        }
      />

      <Route path="/" element={<Navigate to="/produtos" replace />} />
      <Route path="*" element={<Navigate to="/produtos" replace />} />
    </Routes>
  );
}

export default App;
