angular.module('demo', [])
.controller('Hello', function($scope, $http) {
  $scope.page = 'local'
  $scope.local = {}
  $scope.peers = []
  $scope.values = {}
  $scope.providers = {}

  console.log($scope)
  $http.get('/api/v1/local/').then(function(response) {
    $scope.local.id = response.data.id
    console.log('local', $scope.local)
  })
  $http.get('/api/v1/peers/').then(function(response) {
    $scope.peers.length = 0
    for (var i = 0; i < response.data.length; i++) {
      $scope.peers.push(response.data[i])
    }
    console.log('peers', $scope.peers)
  })
  $http.get('/api/v1/values/').then(function(response) {
    for (var prop in $scope.values) {
      $scope.values[prop] = null
    }
    for (var prop in response.data) {
      $scope.values[prop] = response.data[prop]
    }
    console.log('values', $scope.values)
  })
  $http.get('/api/v1/providers/').then(function(response) {
    for (var prop in $scope.values) {
      $scope.providers[prop] = null
    }
    for (var prop in response.data) {
      $scope.providers[prop] = response.data[prop]
    }
    console.log('providers', $scope.providers)
  })
})