module.exports = function(grunt){

    grunt.initConfig({
       connect:{
           server:{
               options:{
                   port:8001,
                   livereload:true
               },
               dev: {
                   options: {
                       middleware: function (connect) {
                           return [
                               require('connect-livereload')(),
                               checkForDownload,
                               mountFolder(connect, '.tmp'),
                               mountFolder(connect, 'app')
                           ];
                       }
                   }
               }
           }
       },
        watch:{
            files:["js/**/*.js","**/*.html","css/*.css"],
            options:{
                livereload:true
            }

        }
    })

    grunt.loadNpmTasks('grunt-contrib-connect')
    grunt.loadNpmTasks('grunt-contrib-watch')

    // To run live reload, juste use "grunt run"
    grunt.registerTask('run',["connect:server","watch"])
    grunt.registerTask('default',["watch"])
}