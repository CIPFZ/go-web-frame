// src/services/api/poetry.ts
import { request } from '@umijs/max';

// --- 类型定义 (根据后端 Model) ---

export interface Dynasty {
  ID: number;
  name: string;
  sortOrder: number;
  createdAt?: string;
}

export interface Genre {
  ID: number;
  name: string;
  sortOrder: number;
}

export interface Author {
  ID: number;
  name: string;
  dynastyId: number;
  intro?: string;
  lifeStory?: string;
  avatarUrl?: string;
  dynasty?: Dynasty; // 关联对象
  createdAt?: string;
  updatedAt?: string;
}

export interface PoemWork {
  ID: number;
  title: string;
  authorId: number;
  genreId: number;
  content: string;
  translation?: string;
  annotation?: string;
  appreciation?: string;
  viewCount?: number;
  author?: Author;
  genre?: Genre;
}

// --- 1. 朝代 API ---

export async function getDynastyList(params: any) {
  return request('/api/v1/poetry/dynasty/list', { method: 'GET', params });
}

export async function getAllDynasties() {
  return request('/api/v1/poetry/dynasty/all', { method: 'GET' });
}

export async function createDynasty(data: any) {
  return request('/api/v1/poetry/dynasty', { method: 'POST', data });
}

export async function updateDynasty(id: number, data: any) {
  return request(`/api/v1/poetry/dynasty/${id}`, { method: 'PUT', data });
}

export async function deleteDynasty(id: number) {
  return request(`/api/v1/poetry/dynasty/${id}`, { method: 'DELETE' });
}

// --- 2. 体裁 API (类似朝代) ---

export async function getGenreList(params: any) {
  return request('/api/v1/poetry/genre/list', { method: 'GET', params });
}

export async function getAllGenres() {
  return request('/api/v1/poetry/genre/all', { method: 'GET' });
}

export async function createGenre(data: any) {
  return request('/api/v1/poetry/genre', { method: 'POST', data });
}

export async function updateGenre(id: number, data: any) {
  return request(`/api/v1/poetry/genre/${id}`, { method: 'PUT', data });
}

export async function deleteGenre(id: number) {
  return request(`/api/v1/poetry/genre/${id}`, { method: 'DELETE' });
}

// --- 3. 诗人 API ---

export async function getAuthorList(params: any) {
  return request('/api/v1/poetry/author/list', { method: 'GET', params });
}

export async function createAuthor(data: any) {
  return request('/api/v1/poetry/author', { method: 'POST', data });
}

export async function updateAuthor(id: number, data: any) {
  return request(`/api/v1/poetry/author/${id}`, { method: 'PUT', data });
}

export async function deleteAuthor(id: number) {
  return request(`/api/v1/poetry/author/${id}`, { method: 'DELETE' });
}

// --- 4. 诗词作品 API ---

export async function getPoemList(params: any) {
  return request('/api/v1/poetry/poem/list', { method: 'GET', params });
}

export async function createPoem(data: any) {
  return request('/api/v1/poetry/poem', { method: 'POST', data });
}

export async function updatePoem(id: number, data: any) {
  return request(`/api/v1/poetry/poem/${id}`, { method: 'PUT', data });
}

export async function deletePoem(id: number) {
  return request(`/api/v1/poetry/poem/${id}`, { method: 'DELETE' });
}
