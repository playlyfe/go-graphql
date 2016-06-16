var GraphQLSchema = require('graphql').GraphQLSchema;
var GraphQLObjectType = require('graphql').GraphQLObjectType;
var GraphQLString = require('graphql').GraphQLString;

var express = require('express');
var graphqlHTTP = require('express-graphql');

var app = express();

var MyGraphQLSchema = new GraphQLSchema({
	query: new GraphQLObjectType({
		    name: 'RootQueryType',
			fields: {
				hello: {
					        type: GraphQLString,
							resolve() {
								          return 'world';
										          
							}
							      
				}
				    
			}
			  
	})

});

app.use('/graphql', graphqlHTTP({ schema: MyGraphQLSchema, graphiql: true  }));

app.listen(3002, function () {
	  console.log('Benchmark app listening on port 3002!');

});
