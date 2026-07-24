import { Navigate, Route, Routes } from 'react-router-dom'
import App from './App'
import OrderEditPage from './pages/OrderEditPage'
import OrderListPage from './pages/OrderListPage'
import OrderPrintPage from './pages/OrderPrintPage'

export function AppRoutes() {
  return (
    <Routes>
      <Route element={<App />}>
        <Route index element={<Navigate to="/orders" replace />} />
        <Route path="orders" element={<OrderListPage />} />
        <Route path="orders/new" element={<OrderEditPage />} />
        <Route path="orders/:id" element={<OrderEditPage />} />
        <Route path="orders/:id/print" element={<OrderPrintPage />} />
      </Route>
    </Routes>
  )
}
