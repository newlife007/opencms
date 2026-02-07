import request from '@/utils/request'

/**
 * Search API
 */

// Search files with filters
export function search(params) {
  return request({
    url: '/search',
    method: 'get',
    params: {
      q: params.q,
      type: params.types,  // Array of file types
      status: params.statuses,  // Array of statuses
      category_id: params.category_id,
      date_from: params.date_from,
      date_to: params.date_to,
      page: params.page || 1,
      page_size: params.page_size || 20,
      sort_by: params.sort_by || 'relevance',
    },
  })
}

// Get search suggestions (for autocomplete)
export function getSuggestions(query) {
  return request({
    url: '/search/suggestions',
    method: 'get',
    params: { q: query },
  })
}

// Get search index status (admin only)
export function getIndexStatus() {
  return request({
    url: '/admin/search/status',
    method: 'get',
  })
}

// Trigger search index rebuild (admin only)
export function reindex() {
  return request({
    url: '/admin/search/reindex',
    method: 'post',
  })
}

export default {
  search,
  getSuggestions,
  getIndexStatus,
  reindex,
}
