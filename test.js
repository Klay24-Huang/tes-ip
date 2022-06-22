function TestIp(requestCount = 120, url = 'http://localhost:7071/') {
    let promises = []
    for (let index = 0; index < requestCount; index++) {
        let request = fetch(url)
        promises.push(request)
    }

    Promise.all(promises).then(responses => {
        let okCount = 0
        let toManyRequestCount = 0
        let otherErrorCount = 0
        responses.forEach(response => {
            let statusCode = response.status
            if (statusCode == 200) {
                okCount++
            } else if (statusCode == 429) {
                toManyRequestCount++
            } else {
                otherErrorCount++
            }
        })

        console.log(`
        result: \n
        status 200: ${okCount}, \n
        status 429: ${toManyRequestCount}, \n
        other error: ${otherErrorCount}, \n
        total requests count: ${requestCount}
        `)
    })
}

TestIp()