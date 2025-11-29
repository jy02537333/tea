import React from 'react';
import { BrowserRouter, Routes, Route, Link, Navigate } from 'react-router-dom';
import Dashboard from './pages/Dashboard';
import ProductList from './pages/ProductList';
import Products from './pages/Products';
import OrderDetail from './pages/OrderDetail';
import Users from './pages/Users';
import Login from './pages/Login';

export default function AppRouter() {
  return (
    <BrowserRouter>
      <nav style={{ padding: 12, borderBottom: '1px solid #eee', marginBottom: 16 }}>
        <Link to="/" style={{ marginRight: 16 }}>仪表盘</Link>
        <Link to="/products" style={{ marginRight: 16 }}>商品管理（旧）</Link>
        <Link to="/products-new" style={{ marginRight: 16 }}>商品管理</Link>
        <Link to="/users" style={{ marginRight: 16 }}>用户管理</Link>
        <Link to="/orders/1">订单详情(id=1)</Link>
      </nav>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<Dashboard />} />
        <Route path="/products" element={<ProductList />} />
        <Route path="/products-new" element={<Products />} />
        <Route path="/users" element={<Users />} />
        <Route path="/orders/:id" element={<OrderDetail id={1} />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
