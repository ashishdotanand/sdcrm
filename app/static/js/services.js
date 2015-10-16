var services = angular.module('services', ['ngResource']);

services.factory('BOM', ['$resource', function($resource){
    return $resource('/api/bom', {}, {
      'reset': {method: 'POST', url: 'a/api/bom/reset'},
      'resetToSampleState': {method: 'POST', url: '/a/api/bom/resetToSampleBOM'}
    });
  }]);

services.factory('ORDERS', ['$resource', function($resource){
    return $resource('/api/orders/:id',{},{
    });
  }]);

services.factory('PENDINGORDERS', ['$resource', function($resource){
    return $resource('/api/pendingorders/purchasers/:id',{},{
      'get':{method: 'GET', isArray: true}
    });
  }]);

services.factory('INVOICES', ['$resource', function($resource){
    return $resource('/api/invoices/:id',{},{
      'getExtraItems':{method: 'POST', url: '/api/extraItemsInInvoice/', isArray:false}
    });
  }]);

services.factory('PURCHASERS', ['$resource', function($resource){
    return $resource('/api/purchasers/:id',{},{
    });
  }]);

services.factory('SKUS', ['$resource', function($resource){
    return $resource('/api/skus/purchaser/:id',{},{
    });
  }]);
